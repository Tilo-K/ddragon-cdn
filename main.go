package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/alcortesm/tgz"
	"github.com/otiai10/copy"
)

func CopyDirectory(scrDir, dest string) error {
	entries, err := os.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		fInfo, err := entry.Info()
		if err != nil {
			return err
		}

		isSymlink := fInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, fInfo.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func Copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	defer in.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}

func getVersions() []string {
	url := "https://ddragon.leagueoflegends.com/api/versions.json"

	res, err := http.Get(url)
	checkError(err)
	if err != nil{
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	checkError(err)

	versions := make([]string, 0)
	json.Unmarshal(body, &versions)

	return versions
}

func loadDdragon(version string) string {
	storageDir := os.Getenv("STORAGE_DIR")
	url := fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/dragontail-%s.tgz", version)
	filename := filepath.Join(storageDir, "ddragon-" + version + ".tgz")

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		res, err := http.Get(url)
		checkError(err)
		defer res.Body.Close()

		out, err := os.Create(filename)
		checkError(err)
		defer out.Close()

		io.Copy(out, res.Body)
	}

	return filename
}

func loadRankedEmblems() string {
	storageDir := os.Getenv("STORAGE_DIR")
	url := fmt.Sprintf("https://static.developer.riotgames.com/docs/lol/ranked-emblems.zip")
	filename := filepath.Join(storageDir, "ranked-emblems.zip")

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		res, err := http.Get(url)
		checkError(err)
		defer res.Body.Close()

		out, err := os.Create(filename)
		checkError(err)
		defer out.Close()

		io.Copy(out, res.Body)
	}

	return filename
}

func getCurrentVersion() string {
	storageDir := os.Getenv("STORAGE_DIR")
	_, err := os.Stat(filepath.Join(storageDir, "current.txt"))
	if os.IsNotExist(err) {
		return ""
	}

	_, err = os.Stat(filepath.Join(storageDir, "data"))
	if os.IsNotExist(err) {
		return ""
	}

	data, err := ioutil.ReadFile(filepath.Join(storageDir, "current.txt"))
	checkError(err)

	return string(data)
}

func loadCurrent() {
	storageDir := os.Getenv("STORAGE_DIR")
	versions := getVersions()
	if versions == nil{
		return
	}
	if getCurrentVersion() != versions[0]{
		file := loadDdragon(versions[0])
		ioutil.WriteFile(filepath.Join(storageDir, "current.txt"), []byte(versions[0]), 0777)

		path, err := tgz.Extract(file)
		checkError(err)

		_, err = os.Stat(filepath.Join(storageDir, "data"))
		if !os.IsNotExist(err) {
			err = os.RemoveAll(filepath.Join(storageDir, "data"))
		}
		checkError(err)

		dest, err := filepath.Abs(filepath.Join(storageDir, "data"))
		checkError(err)
		err = os.Rename(path, dest)
		if err != nil {
			copy.Copy(path, dest)
		}
		os.RemoveAll(path)

		src, _ := filepath.Abs(filepath.Join(storageDir, "data", versions[0]))
		dst, _ := filepath.Abs(filepath.Join(storageDir, "data", "latest"))

		err = os.Rename(src, dst)
		checkError(err)

		export, _ := filepath.Abs(filepath.Join(storageDir, "data", "ranked-emblems"))
		err = CopyDirectory(filepath.Join(storageDir, "ranked-emblems"), export)
		checkError(err)
	}
}

func cors(fs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		(w).Header().Set("Access-Control-Allow-Origin", "*")

		fs.ServeHTTP(w, r)
	}
}

func main() {
	storageDir := os.Getenv("STORAGE_DIR")
	loadCurrent()

	go func() {
		dur, err := time.ParseDuration("30m")
		checkError(err)

		for true {
			time.Sleep(dur)
			loadCurrent()
		}
	}()

	fs := http.FileServer(http.Dir(filepath.Join(storageDir, "data")))
	http.Handle("/", cors(fs))
	port := os.Getenv("PORT")

	if port == ""{
		port = "60002"
	}

	log.Print("Listening on :" + port + "...")
	err := http.ListenAndServe(":" + port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

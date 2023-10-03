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

func getVersions() []string {
	url := "https://ddragon.leagueoflegends.com/api/versions.json"

	res, err := http.Get(url)
	checkError(err)
	if err != nil {
		return nil
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	checkError(err)

	versions := make([]string, 0)
	json.Unmarshal(body, &versions)

	return versions
}

func loadDdragon(version string) string {
	storageDir := os.Getenv("STORAGE_DIR")
	url := fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/dragontail-%s.tgz", version)
	filename := filepath.Join(storageDir, "ddragon-"+version+".tgz")

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
	deleteTgz()
	storageDir := os.Getenv("STORAGE_DIR")
	versions := getVersions()
	if versions == nil {
		return
	}
	if getCurrentVersion() != versions[0] {
		file := loadDdragon(versions[0])
		os.WriteFile(filepath.Join(storageDir, "current.txt"), []byte(versions[0]), 0777)

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
		err = CopyDirectory("ranked-emblems", export)
		checkError(err)
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

	if port == "" {
		port = "60002"
	}

	log.Print("Listening on :" + port + "...")
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

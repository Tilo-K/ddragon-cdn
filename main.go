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
	"github.com/xyproto/unzip"
)

func getVersions() []string {
	url := "https://ddragon.leagueoflegends.com/api/versions.json"

	res, err := http.Get(url)
	checkError(err)
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	checkError(err)

	versions := make([]string, 0)
	json.Unmarshal(body, &versions)

	return versions
}

func loadDdragon(version string) string {
	url := fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/dragontail-%s.tgz", version)
	filename := "ddragon-" + version + ".tgz"

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
	url := fmt.Sprintf("https://static.developer.riotgames.com/docs/lol/ranked-emblems.zip")
	filename := "ranked-emblems.zip"

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
	_, err := os.Stat("current.txt")
	if os.IsNotExist(err) {
		return ""
	}

	_, err = os.Stat("data")
	if os.IsNotExist(err) {
		return ""
	}

	data, err := ioutil.ReadFile("current.txt")
	checkError(err)

	return string(data)
}

func loadCurrent() {
	versions := getVersions()
	if getCurrentVersion() != versions[0] {
		file := loadDdragon(versions[0])
		ioutil.WriteFile("current.txt", []byte(versions[0]), 0777)

		path, err := tgz.Extract(file)
		checkError(err)

		_, err = os.Stat("data")
		if !os.IsNotExist(err) {
			err = os.RemoveAll("data")
		}
		checkError(err)

		dest, err := filepath.Abs("./data")
		checkError(err)
		err = os.Rename(path, dest)
		if err != nil {
			copy.Copy(path, dest)
		}
		os.RemoveAll(path)

		src, _ := filepath.Abs(filepath.Join("data", versions[0]))
		dst, _ := filepath.Abs(filepath.Join("data", "latest"))

		err = os.Rename(src, dst)
		checkError(err)

		emblemsFile := loadRankedEmblems()
		export, _ := filepath.Abs(filepath.Join("data", "ranked-emblems"))
		err = unzip.Extract(emblemsFile, export)
		checkError(err)

		os.RemoveAll(emblemsFile)
	}
}

func cors(fs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		(w).Header().Set("Access-Control-Allow-Origin", "*")

		fs.ServeHTTP(w, r)
	}
}

func main() {
	loadCurrent()

	go func() {
		dur, err := time.ParseDuration("30m")
		checkError(err)

		for true {
			time.Sleep(dur)
			loadCurrent()
		}
	}()

	fs := http.FileServer(http.Dir("./data"))
	http.Handle("/", cors(fs))

	log.Print("Listening on :60002...")
	err := http.ListenAndServe(":60002", nil)
	if err != nil {
		log.Fatal(err)
	}
}

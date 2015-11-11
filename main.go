package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"
)

func log(f *os.File, message string) {
	f.WriteString(fmt.Sprintf("%s\n", message))
	f.Sync()
}

func main() {
	if len(os.Args) < 7 {
		return
	}

	anime := os.Args[1]
	episode := os.Args[2]
	url := os.Args[3]
	filename := os.Args[4]
	outdir := os.Args[5]
	intermdir := os.Args[6]

	if filename[len(filename)-4:] != ".mkv" {
		filename += ".mkv"
	}

	os.Mkdir(".werker-logs", 0755)

	flog, err := os.Create(fmt.Sprintf(".werker-logs/werker-nyaa-%s-%s_%s.log", anime, episode, time.Now().Format("2006-01-02-15-04-05")))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer flog.Close()

	log(flog, "starting werker\n")
	log(flog, fmt.Sprintf("url: %s", url))

	resp, err := http.Get(url)
	if err != nil {
		log(flog, fmt.Sprintf("error downloading: %s", err))
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log(flog, fmt.Sprintf("error reading response: %s", err))
		return
	}

	err = ioutil.WriteFile(path.Join(intermdir, ".watch", fmt.Sprintf("%s.torrent", filename)), data, 0644)
	if err != nil {
		log(flog, fmt.Sprintf("error writing torrent: %s", err))
		return
	}
	log(flog, "wrote torrent to watch directory. creating dummy file in output directory")

	os.Mkdir(fmt.Sprintf("%s/%s", outdir, anime), 0755)
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s/%s - %s.mkv", outdir, anime, anime, episode), []byte{}, 0666)
	if err != nil {
		log(flog, fmt.Sprintf("could not create dummy file: %s", err))
		return
	}
	log(flog, "dummy file created")
	log(flog, "starting watchdog to wait for finished MKV")

	ticker := time.NewTicker(time.Second * 10)

	for _ = range ticker.C {
		// check whether the file exists
		stat, err := os.Stat(path.Join(intermdir, "finished", filename))
		if err != nil {
			fmt.Println(err)
			continue
		}
		if time.Since(stat.ModTime()) < (10 * time.Second) {
			fmt.Println("modtime")
			continue
		}
		log(flog, "seems like the torrent finished downloading.")
		log(flog, "moving mkv to final directory")

		err = os.Rename(path.Join(intermdir, "finished", filename), fmt.Sprintf("%s/%s/%s - %s.mkv", outdir, anime, anime, episode))
		if err != nil {
			log(flog, fmt.Sprintf("could not move file: %s"))
			return
		}

		log(flog, "removing torrent file")
		err = os.Remove(path.Join(intermdir, ".watch", fmt.Sprintf("%s.torrent", filename)))
		if err != nil {
			log(flog, fmt.Sprintf("could not remove file: %s"))
			return
		}
		log(flog, "done")
		break
	}
}

package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/calmh/github"
)

func main() {
	downloaders := 8
	project := "syncthing/syncthing"
	dir := ""

	flag.IntVar(&downloaders, "dl", downloaders, "Number of parallel downloads")
	flag.StringVar(&project, "project", project, "Project to download")
	flag.StringVar(&dir, "dir", dir, "Destination directory")
	flag.Parse()

	log.SetOutput(os.Stdout)

	if dir != "" {
		if err := os.Chdir(dir); err != nil {
			log.Fatal(err)
		}
	}

	rels, err := github.LoadReleases(project)
	if err != nil {
		log.Fatal(err)
	}

	orders := make(chan order)
	var wg sync.WaitGroup
	for i := 0; i < downloaders; i++ {
		wg.Add(1)
		go func() {
			downloader(orders)
			wg.Done()
		}()
	}

	for _, rel := range rels {
		for _, asset := range rel.Assets {
			name := filepath.Join(rel.TagName, path.Base(asset.Name))
			if _, err := os.Stat(name); err == nil {
				continue
			}
			orders <- order{
				tag: rel.TagName,
				url: asset.BrowserDownloadURL,
			}
		}
	}

	close(orders)
	wg.Wait()
}

type order struct {
	tag string
	url string
}

func downloader(orders <-chan order) {
	for o := range orders {
		log.Println("Downloading", path.Join(o.tag, path.Base(o.url)), "...")
		if err := downloadReleaseAsset(o.tag, o.url); err != nil {
			log.Println("Download failed:", err)
		}
	}
}

func downloadReleaseAsset(tag, url string) error {
	if _, err := os.Stat(tag); err != nil {
		if err := os.Mkdir(tag, 0777); err != nil && !os.IsExist(err) {
			return err
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	defer resp.Body.Close()

	filename := filepath.Join(tag, path.Base(url))
	if _, err := os.Stat(filename); err == nil {
		return nil
	}

	tmpname := filename + ".tmp"
	defer os.Remove(tmpname)
	out, err := os.Create(tmpname)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return os.Rename(tmpname, filename)
}

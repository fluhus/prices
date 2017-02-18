package scrapers

// A scraper for Eden Teva Market.

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// Homepage for file list.
const edenHome = "http://operations.edenteva.co.il/Prices/index"

// Prefix of download URLs.
const edenFile = "http://operations.edenteva.co.il/Prices/"

// Scrapes data from Eden Teva Market.
type edenScraper struct{}

// Returns a new Eden Teva Market scraper.
func Eden() Scraper {
	return &edenScraper{}
}

func (a *edenScraper) Scrape(dir string) error {
	// Create output directory.
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return fmt.Errorf("Failed to make dir: %v", err)
	}

	fileList, err := a.fileList()
	if err != nil {
		return fmt.Errorf("Failed to get file list: %v", err)
	}

	// Start pusher thread.
	files := make(chan string)
	done := make(chan error)
	go func() {
		for _, file := range fileList {
			files <- file
		}
		close(files)
	}()

	// Start downloader threads.
	for i := 0; i < numberOfThreads; i++ {
		go func() {
			for file := range files {
				_, err := downloadIfNotExists(edenFile+file,
					filepath.Join(dir, file), nil)
				if err != nil {
					done <- err
					return
				}
			}

			done <- err
		}()
	}

	// Wait for downloaders.
	for i := 0; i < numberOfThreads; i++ {
		e := <-done
		if e != nil {
			err = e
		}
	}

	// Drain pusher thread.
	for range files {
	}

	return nil
}

// Returns a list of all files in Eden's page.
func (a *edenScraper) fileList() ([]string, error) {
	// Get homepage.
	res, err := http.Get(edenHome)
	if err != nil {
		return nil, fmt.Errorf("Failed to get homepage: %v", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read homepage: %v", err)
	}

	// Parse links.
	result := []string{}
	files := regexp.MustCompile("<a href=\"(.*?)\"").FindAllSubmatch(body, -1)
	if len(files) == 0 {
		return nil, fmt.Errorf("Got 0 files.")
	}

	for _, file := range files {
		// All links should end with '.zip'. A change in that condition means
		// that the homepage had changed.
		if !bytes.HasSuffix(file[1], []byte(".zip")) {
			return nil, fmt.Errorf("Found a link that's not a zip file: %s",
				file[1])
		}

		result = append(result, string(file[1]))
	}

	return result, nil
}

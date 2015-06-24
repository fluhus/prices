package aggregators

// An aggregator for Yeinot Bitan.

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"regexp"
	"os"
	"path/filepath"
)

// Homepage for file list.
const bitanHome = "http://www.ybitan.co.il/pirce_update"

// Prefix of download URLs.
const bitanFile = "http://www.ybitan.co.il/upload/"

// Aggregates data from Yeinot Bitan.
type bitanAggregator struct {}

// Returns a new Yeinot Bitan aggregator.
func NewBitanAggregator() Aggregator {
	return &bitanAggregator{}
}

func (a *bitanAggregator) Aggregate(dir string) error {
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
	done  := make(chan error)
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
				_, err := downloadIfNotExists(bitanFile + file,
						filepath.Join(dir, file), nil)
				if err != nil {
					done <- err
					return
				}
			}
			
			done <- nil
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
	for range files {}
	
	return err
}

// Returns a list of all files in Bitan's page.
func (a *bitanAggregator) fileList() ([]string, error) {
	// Get homepage.
	res, err := http.Get(bitanHome)
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
	files := regexp.MustCompile(
			"<a href=\"/upload/(.*?\\.zip)\"").FindAllSubmatch(body, -1)
	if len(files) == 0 {
		return nil, fmt.Errorf("Got 0 files.")
	}
	
	for _, file := range files {
		result = append(result, string(file[1]))
	}
	
	return result, nil
}



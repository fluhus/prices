package aggregators

// An aggregator for the Mega chain.

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"regexp"
	"path/filepath"
	"runtime"
	"log"
	"os"
)

// Home page of the mega price site.
const megaHome = "http://publishprice.mega.co.il/"

// An aggregator for the Mega chain.
type megaAggregator struct{}

// Returns a new Shufersal aggregator.
func NewMegaAggregator() Aggregator {
	return &megaAggregator{}
}

func (a *megaAggregator) Aggregate(dir string) error {
	// Create output directory.
	err := os.MkdirAll(dir, 0)
	if err != nil {
		return fmt.Errorf("Failed to make dir: %v", err)
	}
	
	// Start downloader threads.
	numberOfThreads := runtime.NumCPU()
	fileChan := make(chan *dirFile, numberOfThreads)
	doneChan := make(chan error, numberOfThreads)
	
	for i := 0; i < numberOfThreads; i++ {
		go func() {
			for df := range fileChan {
				to := filepath.Join(dir, df.file)
				_, err := downloadIfNotExists(megaHome + df.dir + df.file,
						to, nil)
				if err != nil {
					doneChan <- err
					return
				}
			}
			
			doneChan <- nil
		}()
	}
	
	// Start pusher thread.
	pusherChan := make(chan error, 1)
	go func() {
		// Get files for download.
		dirs, err := a.getDirectories()
		if err != nil {
			pusherChan <- err
			return
		}
		if len(dirs) == 0 {
			close(fileChan)
			pusherChan <- fmt.Errorf("Found 0 directories.")
			return
		}
		log.Printf("Found %d directories.", len(dirs))
		
		for i := range dirs {
			// Get file list.
			log.Printf("Getting file list #%d.", i)
			files, err := a.getFiles(dirs[i])
			if err != nil {
				close(fileChan)
				pusherChan <- err
				return
			}
			if len(files) == 0 {
				close(fileChan)
				pusherChan <- fmt.Errorf("Found 0 files in %s.", dirs[i])
				return
			}
			
			// Push to downloader threads.
			for _, file := range files {
				fileChan <- &dirFile{dirs[i], file}
			}
		}
		
		close(fileChan)
		pusherChan <- nil
	}()
	
	// Wait for threads to finish (including pusher thread).
	for i := 0; i < numberOfThreads; i++ {
		e := <- doneChan
		if e != nil {
			err = e
		}
	}
	
	for range fileChan {}
	e := <-pusherChan
	if e != nil {
		err = e
	}

	return err
}

// Returns paths of subdirectories of the price page.
func (a *megaAggregator) getDirectories() ([]string, error) {
	// Get home page.
	res, err := http.Get(megaHome)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get page (status %s).", res.Status)
	}
	
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	
	// Parse directory names.
	re := regexp.MustCompile("<a href=\"(2\\d{7}/)\"")
	match := re.FindAllSubmatch(body, -1)
	
	if len(match) == 0 {
		return nil, fmt.Errorf("Found no directories.")
	}
	
	dirs := make([]string, len(match))
	for i := range match {
		dirs[i] = string(match[i][1])
	}
	
	return dirs, nil
}

// Returns paths of files in a subdirectory. The paths are ready for download.
// dir should be as returned from getDirectories.
func (a *megaAggregator) getFiles(dir string) ([]string, error) {
	// Get home page.
	res, err := http.Get(megaHome + dir)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get page (status %s).", res.Status)
	}
	
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Parse file names.
	re := regexp.MustCompile("<a href=\"([^\"]*\\.gz)\"")
	match := re.FindAllSubmatch(body, -1)
	
	if len(match) == 0 {
		return nil, fmt.Errorf("Found no files.")
	}
	
	files := make([]string, len(match))
	for i := range match {
		files[i] = string(match[i][1])
	}
	
	return files, nil
}

// A directory and a file. Surprised? So are we!
type dirFile struct {
	dir string
	file string
}

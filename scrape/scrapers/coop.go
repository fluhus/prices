package scrapers

// A scraper for the Co-Op chain.

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	urllib "net/url"
	"os"
	"path/filepath"
	"regexp"
)

type coopScraper struct{}

// Returns a new Co-Op scraper.
func Coop() Scraper {
	return &coopScraper{}
}

func (a *coopScraper) Scrape(dir string) error {
	// Get files for download.
	infos, infosDone := a.filesForDownload()

	// Start downloader threads.
	done := make(chan error, numberOfThreads)
	for i := 0; i < numberOfThreads; i++ {
		go func() {
			for info := range infos {
				err := a.download(info.url, dir, info.values)
				if err != nil {
					log.Printf("Download error for '%v', branch %v: %v",
						info.url, info.values["branch"][0], err)
					continue
				}
			}

			done <- nil
		}()
	}

	// Wait for downloaders to finish.
	var err error
	for i := 0; i < numberOfThreads; i++ {
		e := <-done
		if e != nil {
			err = e
		}
	}

	// Drain pusher.
	for range infos {
	}
	e := <-infosDone
	if e != nil {
		err = e
	}

	return err
}

// Information for a single download.
type coopFileInfo struct {
	url    string
	values urllib.Values
}

// Returns a channel that will yield file-infos for download. The error channel
// will report when it's finished.
func (a *coopScraper) filesForDownload() (chan *coopFileInfo, chan error) {
	// Instantiate channels.
	infos := make(chan *coopFileInfo, numberOfThreads)
	done := make(chan error, 1)

	// Start file getter thread.
	go func() {
		// Make sure to close everythins.
		var err error
		defer close(done)
		defer close(infos)
		defer func() { done <- err }()

		// Get stores file.
		infos <- &coopFileInfo{
			"http://coopisrael.coop/home/branches_to_xml",
			form("type", "1", "agree", "1"),
		}

		// Get branches.
		res, err := httpPost(
			"http://coopisrael.coop/ajax/search_branch", nil, nil)
		if err != nil {
			err = fmt.Errorf("Failed to request branches: %v", err)
			return
		}
		if res.StatusCode != http.StatusOK {
			err = fmt.Errorf("Failed to request branches: %s", res.Status)
			return
		}

		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			err = fmt.Errorf("Failed to request branches: %v", err)
			return
		}

		// Extract branch IDs.
		branchesRaw := regexp.MustCompile("data-id=\\\\\"(.+?)\\\\\"").
			FindAllSubmatch(body, -1)
		if len(branchesRaw) == 0 {
			err = fmt.Errorf("Found 0 branches.")
			return
		}
		branches := make([]string, len(branchesRaw))
		for i := range branchesRaw {
			branches[i] = string(branchesRaw[i][1])
		}

		log.Printf("Found %d branches.", len(branches))

		// Push promos & prices.
		for _, branch := range branches {
			infos <- &coopFileInfo{
				"http://coopisrael.coop/home/get_promo",
				form("branch", branch, "type", "1", "agree", "1"),
			}
			infos <- &coopFileInfo{
				"http://coopisrael.coop/home/get_prices",
				form("branch", branch, "type", "1", "agree", "1",
					"product", "0"),
			}
		}
	}()

	return infos, done
}

// Downloads a given file from Co-Op.
func (a *coopScraper) download(url, dir string, values urllib.Values) error {
	// Open connection to site.
	res, err := httpPost(url, values, nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to request file: Got status %s.", res.Status)
	}

	// Extract file name.
	if len(res.Header["Content-Disposition"]) == 0 {
		return fmt.Errorf("No file name in response.")
	}
	fileNameRaw := res.Header["Content-Disposition"][0]
	fileName := string(find([]byte(fileNameRaw), "filename=([^;]*)"))
	if fileName == "" {
		return fmt.Errorf("No file name in response.")
	}
	fileName += ".gz"
	to := filepath.Join(dir, fileName)

	log.Printf("Downloading '%s' to '%s'.", url, to)

	// Open output file.
	fout, err := os.Create(to)
	if err != nil {
		return fmt.Errorf("Failed to create output file: %v", err)
	}
	defer fout.Close()
	bout := bufio.NewWriter(fout)
	defer bout.Flush()
	zout := gzip.NewWriter(bout)
	defer zout.Close()

	// Download!
	_, err = io.Copy(zout, res.Body)

	return err
}

// Creates a values object for POST requests. Arguments are pairs of key and
// value.
func form(s ...string) urllib.Values {
	if len(s)%2 != 0 {
		panic("Got odd number of arguments.")
	}

	result := map[string][]string{}
	for i := 0; i < len(s); i += 2 {
		result[s[i]] = []string{s[i+1]}
	}

	return urllib.Values(result)
}

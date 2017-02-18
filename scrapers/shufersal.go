package scrapers

// A scraper for the Shufersal chain.

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"regexp"
	"strconv"
	"path/filepath"
	"log"
	"os"
	"html"
)

// A scraper for the Shufersal chain.
type shufersalScraper struct{}

// Returns a new Shufersal scraper.
func Shufersal() Scraper {
	return &shufersalScraper{}
}

func (a *shufersalScraper) Scrape(dir string) error {
	// Create output directory.
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return fmt.Errorf("Failed to make dir: %v", err)
	}

	// Get number of pages from the first page.
	page, err := a.getPage(1)
	if err != nil {
		return fmt.Errorf("Failed to get page 1: %v", err)
	}
	
	numberOfPages := a.parseLastPageNumber(page)
	if numberOfPages == -1 {
		return fmt.Errorf("Failed to parse number of pages.")
	}
	log.Printf("Parsing %d pages.", numberOfPages)
	
	// Download!
	numChan := make(chan int, numberOfThreads)
	done := make(chan error, numberOfThreads)
	
	for i := 0; i < numberOfThreads; i++ {
		go func() {
			for i := range numChan {
				// Parse page.
				log.Printf("Parsing page %d.", i)
				page, err := a.getPage(i)
				if err != nil {
					done <- err
					return
				}
				
				entries, err := a.parsePage(page)
				if err != nil {
					done <- err
					return
				}
				log.Printf("Page %d has %d entries.", i, len(entries))
				
				// Download entries.
				for _, entry := range entries {
					to := filepath.Join(dir, entry.file)
					_, err := downloadIfNotExists(entry.url, to, nil)
					if err != nil {
						done <- err
						return
					}
				}
			}
			
			done <- nil
		}()
	}
	
	// Give page numbers.
	go func() {
		for i := 1; i <= numberOfPages; i++ {
			numChan <- i
		}
		close(numChan)
	}()
	
	// Join threads.
	for i := 0; i < numberOfThreads; i++ {
		e := <-done
		if e != nil {
			err = e
		}
	}
	
	// Drain num channel.
	for range numChan {}

	return err
}

// Returns the body of the n'th page in Shufersal's site.
func (a *shufersalScraper) getPage(n int) ([]byte, error) {
	res, err := http.Get(fmt.Sprintf("http://prices.shufersal.co.il/?page=%d",
			n))
	if err != nil { return nil, err }
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bad response status: %s", res.Status)
	}
	
	return ioutil.ReadAll(res.Body)
}

// A single downloadable file.
type shufersalEntry struct {
	index int
	url   string
	file  string
}

// Parses entries out of the given page body. Returns an error if parsing fails.
func (a *shufersalScraper) parsePage(page []byte) ([]*shufersalEntry,
		error) {
	result := []*shufersalEntry{}
	
	// Some regular expressions.
	pageSplitter := regexp.MustCompile("(?s)<tr.*?</tr>")
	rowSplitter := regexp.MustCompile("(?s)<td>(.*?)</td>")
	urlGetter := regexp.MustCompile("^<a href=\"(.*?)\"")
	fileGetter := regexp.MustCompile("^[^?]*/([^/?]*)?")
	
	rows := pageSplitter.FindAll(page, -1)
	if len(rows) < 3 {  // Should have at least one entry.
		return nil, fmt.Errorf("Bad number of rows: %d, expected at least 3.",
				len(rows))
	}
	
	rows = rows[2:]  // Skip header and footer.
	for _, row := range rows {
		// Split row to columns.
		fields := rowSplitter.FindAllSubmatch(row, -1)
		if len(fields) != 8 {
			return nil, fmt.Errorf("Bad number of fields: %d, expected 8.",
					len(fields))
		}
		
		// Parse fields.
		index, err := strconv.Atoi(string(fields[7][1]))
		if err != nil {
			return nil, fmt.Errorf("Cannot parse entry index: %s", fields[7][1])
		}
		url := urlGetter.FindSubmatch(fields[0][1])
		if url == nil {
			return nil, fmt.Errorf("cannot not find url in: %s", fields[0][1])
		}
		file := fileGetter.FindSubmatch(url[1])
		if file == nil {
			return nil, fmt.Errorf("cannot not find file name in: %s", url[1])
		}
		
		// Append new entry.
		entry := &shufersalEntry{}
		entry.index = index
		entry.url = html.UnescapeString(string(url[1]))
		entry.file = string(file[1])
		result = append(result, entry)
	}
	
	return result, nil
}

// Returns the number of the last page, or -1 if failed to parse.
func (a *shufersalScraper) parseLastPageNumber(page []byte) int {
	// log.Print(string(page))
	re := regexp.MustCompile("<a [^>]*?href=\"/\\?page=(\\d+)")
	
	match := re.FindAllSubmatch(page, -1)
	if len(match) == 0 {
		return -1
	}
	
	num, err := strconv.Atoi(string(match[len(match) - 1][1]))
	if err != nil {
		return -1
	}
	
	return num
}

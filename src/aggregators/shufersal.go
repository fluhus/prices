package aggregators

// An aggregator for the Shufersal chain.

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"regexp"
	"strconv"
)

type ShufersalAggregator struct{}

// Returns a new Shufersal aggregator.
func NewShufersalAggregator() *ShufersalAggregator {
	return &ShufersalAggregator{}
}

func (a *ShufersalAggregator) Aggregate(dir string) error {
	page, err := a.getPage(1)
	if err != nil { return err }
	_, err = a.parsePage(page)
	if err != nil { return err }
	// fmt.Println(string(page))

	return nil
}

// Returns the body of the n'th page in Shufersal's site.
func (a *ShufersalAggregator) getPage(n int) ([]byte, error) {
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
func (a *ShufersalAggregator) parsePage(page []byte) ([]*shufersalEntry,
		error) {
	result := []*shufersalEntry{}
	
	// Some regular expressions.
	pageSplitter := regexp.MustCompile("(?s)<tr.*?</tr>")
	rowSplitter := regexp.MustCompile("(?s)<td>(.*?)</td>")
	urlGetter := regexp.MustCompile("^<a href=\"(.*?)\"")
	fileGetter := regexp.MustCompile("^[^?]*/([^/?]*)?")
	
	rows := pageSplitter.FindAll(page, -1)
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
		entry.url = string(url[1])
		entry.file = string(file[1])
		result = append(result, entry)
	}
	
	return result, nil
}

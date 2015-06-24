package aggregators

// An aggregator for Nibit-based chains.

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"regexp"
	"time"
	//"io"
	//"os"
	//"path/filepath"
	"log"
	urllib "net/url"
)

const (
	nibitHome = "http://matrixcatalog.co.il/"                // For cookies.
	nibitPage = nibitHome + "NBCompetitionRegulations.aspx"  // For queries.
)

// Chain ID's for filtering.
const (
	Victory = "7290696200003"
	Hashook = "7290661400001"
	Lahav   = "7290058179503"
)

// Aggregates data from Nibit.
type nibitAggregator struct {
	chain string  // Name of chain.
}

// Returns a new Nibit aggregator.
func NewNibitAggregator(chain string) Aggregator {
	return &nibitAggregator{chain}
}

func (a *nibitAggregator) Aggregate(dir string) error {
	// Create output directory.
	//err := os.MkdirAll(dir, 0700)
	//if err != nil {
	//	return fmt.Errorf("Failed to make dir: %v", err)
	//}
	
	log.Print("Starting session.")
	cl, err := a.startSession()
	if err != nil {
		return fmt.Errorf("Failed to start session: %v", err)
	}

	log.Print("Getting file names.")
	files, err := a.fileList(cl, time.Now())
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("Found 0 files.")
	}
	fmt.Println(len(files))
	
	return nil
}

// Returns a client with a session ID cookie.
func (a *nibitAggregator) startSession() (*http.Client, error) {
	// Get homepage.
	res, err := http.Head(nibitPage)
	if err != nil {
		return nil, fmt.Errorf("Failed to request homepage: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to request homepage: Bad response " +
				"status: %s", res.Status)
	}
	
	// Parse cookie.
	name, value := a.parseSessionCookie(res)
	if name == "" || value == "" {
		return nil, fmt.Errorf("Failed to get session ID.")
	}
	
	return &http.Client{Jar: singleCookieJar(nibitHome, name, value)}, nil
}

// Parses the session ID cookie of the given response. Returns empty strings
// if not found.
func (a *nibitAggregator) parseSessionCookie(res *http.Response) (name,
		value string) {
	// Check for Set-Cookie field.
	if len(res.Header["Set-Cookie"]) == 0 {
		return "", ""
	}
	
	match := regexp.MustCompile("^([^;=]+)=([^;]+)").FindStringSubmatch(
			res.Header["Set-Cookie"][0])
	if match == nil {
		return "", ""
	}
	return match[1], match[2]
}

// Downloads a file list for the given date and chain. Chain can be -1 for all
// chains.
func (a *nibitAggregator) fileList(cl *http.Client, date time.Time) (
		files []string, err error) {
	// Get homepage.
	res, err := cl.Get(nibitPage)
	if err != nil {
		return nil, fmt.Errorf("Failed to read page: %v", err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Failed to read page body: %v", err)
	}
	
	// Go over pages.
	pageNumber := 1
	for {
		log.Printf("Parsing page %d.", pageNumber)
	
		// Update form values.
		values := a.formValues(body)
		a.setFormDate(values, date)
		a.setFormChain(values, a.chain)
		if pageNumber == 1 {
			a.setFormActionSearch(values)
		} else {
			a.setFormActionNext(values)
		}
		pageNumber++
		
		// Send post.
		res, err = cl.PostForm(nibitPage, values)
		if err != nil {
			return nil, fmt.Errorf("Failed to read page: %v", err)
		}
		body, err = ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("Failed to read page body: %v", err)
		}
		
		// Parse table.
		rowsRe := regexp.MustCompile("(?s)<tr>(.*?)</tr>")
		colsRe := regexp.MustCompile("(?s)<td>(.*?)</td>")
		
		rows := rowsRe.FindAllSubmatch(body, -1)
		if len(rows) == 0 {
			return nil, fmt.Errorf("Found 0 files on page %d.", pageNumber)
		}
		
		for _, row := range rows {
			// Break into columns.
			cols := colsRe.FindAllSubmatch(row[1], -1)
			if len(cols) == 0 { continue }  // Maybe a header.
			
			files = append(files, string(cols[0][1]))
		}
		
		// Break if last page.
		if !a.pageHasNext(body) {
			break
		}
	}
	
	return
}

// Parses form values from the given response body. Before using the result for
// a POST request, make sure to set the date and action values.
func (a *nibitAggregator) formValues(body []byte) urllib.Values {
	re := regexp.MustCompile("<input type=\"hidden\" name=\"[^\"]*\" " +
			"id=\"([^\"]*)\" value=\"([^\"]*)\"")
	match := re.FindAllSubmatch(body, -1)
	
	// Get hidden form values.
	result := map[string][]string{}
	for _, m := range match {
		result[string(m[1])] = []string{string(m[2])}
	}
	
	// Add visible form values.
	result["ctl00$txtSearchProduct"] = []string{""}
	result["ctl00$TextArea"] = []string{""}
	result["ctl00$MainContent$chain"] = []string{"-1"}
	result["ctl00$MainContent$subChain"] = []string{"-1"}
	result["ctl00$MainContent$branch"] = []string{"-1"}
	result["ctl00$MainContent$fileType"] = []string{"all"}

	return urllib.Values(result)
}

// Returns a string representation of the given time, as a date for Nibit form.
func (a *nibitAggregator) formatDate(t time.Time) string {
	return fmt.Sprintf("%02d/%02d/%d", t.Day(), t.Month(), t.Year())
}

// Sets the date field of the form.
func (a *nibitAggregator) setFormDate(values urllib.Values, date time.Time) {
	values["MainContent$txtDate"] = []string{a.formatDate(date)}
}

// Sets action to 'search'.
func (a *nibitAggregator) setFormActionSearch(values urllib.Values) {
	delete(values, "ctl00$MainContent$btnNext1")
	delete(values, "ctl00$MainContent$btnPrev1")
	values["ctl00$MainContent$btnSearch"] = []string{"חיפוש"}
}

// Sets action to 'next'.
func (a *nibitAggregator) setFormActionNext(values urllib.Values) {
	delete(values, "ctl00$MainContent$btnSearch")
	delete(values, "ctl00$MainContent$btnPrev1")
	values["ctl00$MainContent$btnNext1"] = []string{"קדימה"}
}

// Sets action to 'prev'.
func (a *nibitAggregator) setFormActionPrev(values urllib.Values) {
	delete(values, "ctl00$MainContent$btnNext1")
	delete(values, "ctl00$MainContent$btnSearch")
	values["ctl00$MainContent$btnPrev1"] = []string{"אחורה"}
}

// Sets the chain field of the form. Give "-1" for all chains.
func (a *nibitAggregator) setFormChain(values urllib.Values,
		chain string) {
	values["ctl00$MainContent$chain"] = []string{chain}
}

// Checks whether the 'next' button is enabled.
func (a *nibitAggregator) pageHasNext(body []byte) bool {
	re := regexp.MustCompile("<input [^>]* name=\"ctl00\\$MainContent\\$btnNext1\" [^>]* disabled=\"disabled\"")
	return !re.Match(body)
}


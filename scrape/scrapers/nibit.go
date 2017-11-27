package scrapers

// A scraper for Nibit-based chains.

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	urllib "net/url"
	"path/filepath"
	"regexp"
	"time"
)

const (
	nibitHome     = "http://matrixcatalog.co.il/"               // For cookies.
	nibitPage     = nibitHome + "NBCompetitionRegulations.aspx" // For queries.
	nibitDownload = "http://matrixcatalog.co.il/" +
		"CompetitionRegulationsFiles/latest/" // For downloads.
)

// Chain ID's for filtering.
const (
	Victory = "7290696200003"
	Hashook = "7290661400001"
	Lahav   = "7290058179503"
)

// Scrapes data from Nibit.
type nibitScraper struct {
	chain string // Name of chain.
	days  int    // How many days from now back it should download.
}

// Returns a new Nibit scraper. Chain is an ID. Days is how many days back
// from today it should download. days=1 means today only, days=2 means today
// and yesterday, etc. A value lesser than 1 will cause a panic.
func Nibit(chain string, days int) Scraper {
	// Check days.
	if days < 1 {
		panic(fmt.Sprintf("Bad number of days: %d. Must be positive.", days))
	}

	return &nibitScraper{chain, days}
}

func (a *nibitScraper) Scrape(dir string) error {
	log.Print("Starting session.")
	cl, err := a.startSession()
	if err != nil {
		return fmt.Errorf("Failed to start session: %v", err)
	}

	for i := 0; i < a.days; i++ {
		date := a.formatDate(time.Now().AddDate(0, 0, -i*1))
		log.Printf("Downloading files from %s.", date)
		err = a.download(cl, date, dir)
		if err != nil {
			return err
		}
	}

	return nil
}

// Returns a client with a session ID cookie.
func (a *nibitScraper) startSession() (*http.Client, error) {
	// Get homepage.
	res, err := http.Head(nibitPage)
	if err != nil {
		return nil, fmt.Errorf("Failed to request homepage: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to request homepage: Bad response "+
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
func (a *nibitScraper) parseSessionCookie(res *http.Response) (name,
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

// Downloads all available files for the given date.
func (a *nibitScraper) download(cl *http.Client, date, dir string) error {
	// Get homepage.
	res, err := httpGet(nibitPage, cl)
	if err != nil {
		return fmt.Errorf("Failed to read page: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return fmt.Errorf("Failed to read page: Got status %s.", res.Status)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return fmt.Errorf("Failed to read page body: %v", err)
	}

	// Update form values.
	values := a.formValues(body)
	a.setFormDate(values, date)
	a.setFormChain(values, a.chain)
	a.setFormActionSearch(values)

	// Send post - to get files for specific date and chain.
	res, err = httpPost(nibitPage, values, cl)
	if err != nil {
		return fmt.Errorf("Failed to read page: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return fmt.Errorf("Failed to read page: Got status %s.", res.Status)
	}
	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return fmt.Errorf("Failed to read page body: %v", err)
	}

	a.clearFormAction(values)

	// Parse table.
	rowsRe := regexp.MustCompile("(?s)<tr>(.*?)</tr>")
	colsRe := regexp.MustCompile("(?s)<td>(.*?)</td>")

	rows := rowsRe.FindAllSubmatch(body, -1)
	if len(rows) == 0 {
		return fmt.Errorf("Found 0 files on page.")
	}
	log.Printf("Found %d rows (including header).", len(rows))
	// (There can be days with no files, so no error for 0 files.)

	infos := make(chan *nibitFileInfo, numberOfThreads)
	done := make(chan error, numberOfThreads)

	// Start pusher thread.
	go func() {
		fileNumber := 1
		// Go over table rows.
		for _, row := range rows {
			// Break into columns.
			cols := colsRe.FindAllSubmatch(row[1], -1)
			if len(cols) == 0 {
				continue
			} // Maybe a header.

			info := &nibitFileInfo{string(cols[0][1]) + ".xml.gz",
				fileNumber}
			fileNumber++

			infos <- info
		}

		close(infos)
	}()

	// Start downloader threads.
	for i := 0; i < numberOfThreads; i++ {
		go func() {
			for info := range infos {
				_, err := downloadIfNotExists(
					nibitDownload+a.chain+"/"+info.name,
					filepath.Join(dir, info.name), cl)
				if err != nil {
					done <- fmt.Errorf("Failed to download '%s': %v",
						info.name, err)
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
	for range infos {
	}

	if err != nil {
		return err
	}

	return nil
}

// Information needed to download a file.
type nibitFileInfo struct {
	name   string // Name by which to save on disk.
	target int    // Event to send to server.
}

// Parses form values from the given response body. Before using the result for
// a POST request, make sure to set the date and action values.
func (a *nibitScraper) formValues(body []byte) urllib.Values {
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
func (a *nibitScraper) formatDate(t time.Time) string {
	return fmt.Sprintf("%02d/%02d/%d", t.Day(), t.Month(), t.Year())
}

// Sets the date field of the form.
func (a *nibitScraper) setFormDate(values urllib.Values, date string) {
	values["ctl00$MainContent$txtDate"] = []string{date}
}

// Sets action to 'search'.
func (a *nibitScraper) setFormActionSearch(values urllib.Values) {
	delete(values, "ctl00$MainContent$btnNext1")
	delete(values, "ctl00$MainContent$btnPrev1")
	values["ctl00$MainContent$btnSearch"] = []string{"חיפוש"}
}

// Removes action.
func (a *nibitScraper) clearFormAction(values urllib.Values) {
	delete(values, "ctl00$MainContent$btnNext1")
	delete(values, "ctl00$MainContent$btnPrev1")
	delete(values, "ctl00$MainContent$btnSearch")
}

// Sets the chain field of the form. Give "-1" for all chains.
func (a *nibitScraper) setFormChain(values urllib.Values,
	chain string) {
	values["ctl00$MainContent$chain"] = []string{chain}
}

// Sets the target for downloading.
func (a *nibitScraper) setFormTarget(values urllib.Values, target int) {
	values["__EVENTTARGET"] = []string{fmt.Sprintf(
		"ctl00$MainContent$repeater$ctl%02d$lblDownloadFile", target)}
}

// Returns a shallow copy of the given values. Keys can be added and removed
// safely.
func (a *nibitScraper) copyValues(values urllib.Values) urllib.Values {
	result := map[string][]string{}
	for m := range values {
		result[m] = values[m]
	}
	return urllib.Values(result)
}

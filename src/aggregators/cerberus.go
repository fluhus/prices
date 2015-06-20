package aggregators

// An aggregator for Cerberus-based databases.

import (
	"encoding/json"
	"path/filepath"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"runtime"
)

const (
	// Cerberus login page.
	cerberusHome = "https://url.publishedprices.co.il/"
	
	// Cerberus user page (with file list).
	cerberusUser = cerberusHome + "login/user"
	
	// File server address.
	cerberusFile = cerberusHome + "file/"
	
	// File download address.
	cerberusDownload = cerberusFile + "d/"
)

// Aggregates data files from the Cerberus server.
type CerberusAggregator struct {
	username string
}

// Returns a new Cerberus aggregator with the given user-name.
func NewCerberusAggregator(username string) *CerberusAggregator {
	return &CerberusAggregator{username}
}

func (a *CerberusAggregator) Aggregate(dir string) error {
	// Create output directory.
	err := os.MkdirAll(dir, 0)
	if err != nil { return fmt.Errorf("Failed to make directory: %v", err) }

	// Login to Cerberus.
	client, err := a.login()
	if err != nil { return fmt.Errorf("Failed to login: %v", err) }
	
	// Download file list.
	files, err := a.getFileList(client)
	if err != nil { return fmt.Errorf("Failed to get file list: %v", err) }
	
	// Filter only data files.
	files = a.filterFileNames(files)
	if len(files) == 0 { return fmt.Errorf("Found no files after filtering.") }
	
	// Download files!
	numberOfThreads := runtime.NumCPU()
	fileChan := make(chan string, numberOfThreads)
	done := make(chan error, numberOfThreads)
	
	// Start downloader threads.
	for i := 0; i < numberOfThreads; i++ {
		go func() {
			for file := range fileChan {
				outFile := filepath.Join(dir, file)
				_, err := downloadIfNotExists(cerberusDownload + file,
						outFile, client)
				if err != nil {
					done <- fmt.Errorf("Failed to download: %v", err)
					return
				}
			}
			done <- nil
		}()
	}
	
	// Push file names to channel.
	for _, file := range files {
		fileChan <- file
	}
	close(fileChan)
	
	// Wait for threads to finish.
	var result error
	for i := 0; i < numberOfThreads; i++ {
		err := <-done
		if err != nil { result = err }
	}
	close(done)
	
	return result
}

// Returns a logged-in client.
func (a *CerberusAggregator) login() (*http.Client, error) {
	// Get login page.
	res, err := http.Get(cerberusHome)
	if err != nil { return nil, fmt.Errorf("Failed to get homepage: %v", err) }
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Got bad response status: %s", res.Status)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil { return nil, fmt.Errorf("Failed to read homepage: %v", err) }
	res.Body.Close()

	// Get token and cookie.	
	token, err := a.parseToken(body)
	if err != nil { return nil, err }
	cookie, err := a.parseCookie(res)
	if err != nil { return nil, err }
	
	// Login!
	jar, err := singleCookieJar(cerberusHome, "cftpSID", string(cookie))
	if err != nil { return nil, err }
	
	cl := &http.Client{Jar: jar}
	res, err = cl.PostForm(
		cerberusUser,
		map[string][]string{
			"csrftoken": []string{string(token)},
			"username": []string{a.username},
			"password": []string{""},
			"Submit": []string{"Sign in"},
		})
	if err != nil { return nil, fmt.Errorf("Failed to post: %v", err) }
	res.Body.Close()
	
	// Get second cookie.
	cookie, err = a.parseCookie(res)
	if err != nil { return nil, err }
	
	// Update client with new cookie.
	cl.Jar, err = singleCookieJar(cerberusHome, "cftpSID", string(cookie))
	if err != nil { return nil, err }
	
	return cl, nil
}

// Parses the Get-Cookie field of a Cerberus response.
func (a *CerberusAggregator) parseCookie(res *http.Response) ([]byte, error) {
	rawCookie, ok := res.Header["Set-Cookie"]
	if !ok {
		return nil, fmt.Errorf("Response does not contain a cookie.")
	}
	cookie := find([]byte(rawCookie[0]), "cftpSID=(.*?);")
	if cookie == nil {
		return nil, fmt.Errorf("Could not parse cookie value. " +
				"Cookie: %v", rawCookie)
	}
	return cookie, nil
}

// Parses the login token from a Cerberus response body.
func (a *CerberusAggregator) parseToken(body []byte) ([]byte, error) {
	token := find(body, "id=\"csrftoken\" value=\"(.*?)\"")
	if token == nil {
		return nil, fmt.Errorf("Could not parse login token.")
	}
	return token, nil
}

// Gets the list of files from Cerberus, using the given logged-in client.
func (a *CerberusAggregator) getFileList(cl *http.Client) ([]string, error) {
	// Request file list.
	res, err := cl.PostForm(cerberusFile + "ajax_dir?sEcho=2&iColumns=5&sColumns=%2C%2C%2C%2C&iDisplayStart=0&iDisplayLength=100000&mDataProp_0=fname&sSearch_0=&bRegex_0=false&bSearchable_0=true&bSortable_0=true&mDataProp_1=type&sSearch_1=&bRegex_1=false&bSearchable_1=true&bSortable_1=false&mDataProp_2=size&sSearch_2=&bRegex_2=false&bSearchable_2=true&bSortable_2=true&mDataProp_3=ftime&sSearch_3=&bRegex_3=false&bSearchable_3=true&bSortable_3=true&mDataProp_4=&sSearch_4=&bRegex_4=false&bSearchable_4=true&bSortable_4=false&sSearch=&bRegex=false&iSortingCols=0&cd=%2F", nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to post request: %v", err)
	}
	
	// Parse file list.
	var resData struct {   // Represents the json response structure.
		AaData []*struct {
			Value string
		}
	}
	err = json.NewDecoder(res.Body).Decode(&resData)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse response: %v", err)
	}
	res.Body.Close()
	
	// An empty file list should be reported.
	if len(resData.AaData) == 0 {
		return nil, fmt.Errorf("Got an empty file list.")
	}
	
	files := make([]string, len(resData.AaData))
	for i := range files {
		files[i] = resData.AaData[i].Value
	}
	
	return files, nil
}

// Returns a slice with only the file names that are relevant for downloading.
func (a *CerberusAggregator) filterFileNames(files []string) []string {
	acceptedPattern := regexp.MustCompile("^((Price|Promo).*gz|Stores.*xml)$")
	result := []string{}
	
	for _, file := range files {
		if acceptedPattern.MatchString(file) {
			result = append(result, file)
		}
	}
	
	return result
}

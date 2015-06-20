// Deals with downloading data files from the different servers, with specific
// configurations for each chain.
package aggregators

// Common utilities for all aggregators.

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"strconv"
	urllib "net/url"
)


// ----- AGGREGATOR TYPE -------------------------------------------------------

// An aggregator downloads data files for a specific chain.
type Aggregator interface {
	Aggregate(dir string) error
}


// ----- COMMON UTILITIES ------------------------------------------------------

// Looks up a regular expression in the given sequence and returns the #1
// captured group. If not found, returns nil - which is different from a 0-long
// array.
func find(text []byte, exp string) []byte {
	return regexp.MustCompile(exp).FindSubmatch(text)[1]
}

// Returns a cookie-jar with a single cookie. Error shouldn't happen unless
// url is malformed.
func singleCookieJar(url, name, value string) (http.CookieJar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil { return nil, err }
	pathUrl, err := urllib.Parse(url)
	if err != nil { return nil, err }
	jar.SetCookies(pathUrl, []*http.Cookie{&http.Cookie{
		Name: name, Value: value}})
	return jar, nil
}

// Returns true iff the given file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Returns the size of the given file. Returns -1 if not exists or error.
func fileSize(path string) int64 {
	stat, err := os.Stat(path)
	if err == nil {
		return stat.Size()
	} else {
		return -1
	}
}

// Returns the content-length field from a response header. Returns -1 if no
// information is available.
func responseSize(res *http.Response) int64 {
	field, ok := res.Header["Content-Length"]
	if !ok { return -1 }
	if len(field) != 1 { return -1 }
	size, err := strconv.ParseInt(field[0], 0, 64)
	if err != nil { return -1 }
	return size
}

// Downloads a file iff the 'to' path does not exist or their sizes differ (for
// aborted downloads). Give a client for logged-in sessions, or nil to start a
// new session. Returns true iff file was downloaded.
func downloadIfNotExists(url, to string, cl *http.Client) (bool, error) {
	// Instantiate client.
	if cl == nil {
		cl = &http.Client{}
	}
	
	// Request header.
	res, err := cl.Head(url)
	if err != nil {
		return false, fmt.Errorf("Failed to request header: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Got bad response status: %s", res.Status)
	}
	res.Body.Close()
	
	// Check if file already exists.
	if fileExists(to) && responseSize(res) != -1 &&
			responseSize(res) == fileSize(to) {
		return false, nil
	}
	
	// Request file.
	res, err = cl.Get(url)
	if err != nil {
		return false, fmt.Errorf("Failed to request file: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Got bad response status: %s", res.Status)
	}
	defer res.Body.Close()
	
	log.Printf("Downloading '%s' to '%s'.", url, to)
	
	// Open output file.
	out, err := os.Create(to)
	if err != nil {
		return false, fmt.Errorf("Failed to create output file: %v", err)
	}
	defer out.Close()
	buf := bufio.NewWriter(out)
	defer buf.Flush()
	
	// Download!
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return false, fmt.Errorf("Failed to save file: %v", err)
	}
	
	return true, nil
}

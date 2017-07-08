// Deals with downloading data files from the different servers, with specific
// configurations for each chain.
package scrapers

// Common utilities for all scrapers.

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	urllib "net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ----- SCRAPER TYPE ---------------------------------------------------------

// A scraper downloads data files for a specific chain.
type Scraper interface {
	// Downloads all available data files into the specified directory. An
	// scraper may decide to omit downloading already existing files. An
	// scraper may use several threads for its work.
	Scrape(dir string) error
}

// ----- COMMON UTILITIES ------------------------------------------------------

// Maximal number of threads to execute on.
var numberOfThreads = runtime.NumCPU()

// Looks up a regular expression in the given sequence and returns the #1
// captured group. If not found, returns nil - which is different from a 0-long
// array.
func find(text []byte, exp string) []byte {
	return regexp.MustCompile(exp).FindSubmatch(text)[1]
}

// Returns a cookie-jar with a single cookie. Error shouldn't happen unless
// url is malformed.
func singleCookieJar(url, name, value string) http.CookieJar {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err.Error())
	}
	pathUrl, err := urllib.Parse(url)
	if err != nil {
		panic(err.Error())
	}
	jar.SetCookies(pathUrl, []*http.Cookie{&http.Cookie{
		Name: name, Value: value}})
	return jar
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
	if !ok {
		return -1
	}
	if len(field) != 1 {
		return -1
	}
	size, err := strconv.ParseInt(field[0], 0, 64)
	if err != nil {
		return -1
	}
	return size
}

// Downloads a file iff the 'to' path does not exist. Give a client for
// logged-in sessions, or nil to start a new session. Returns true iff file was
// downloaded.
func downloadIfNotExists(url, to string, cl *http.Client) (bool, error) {
	to = expandPath(to)
	if ts := fileTimestamp(to); ts == -1 || ts < fromTimestamp {
		return false, nil
	}

	// Create output directory.
	err := mkdir(filepath.Dir(to))
	if err != nil {
		return false, fmt.Errorf("Failed to make dir: %v", err)
	}

	// Instantiate client.
	if cl == nil {
		cl = &http.Client{}
	}

	// Check if file already exists.
	if fileExists(to) && fileSize(to) != 0 {
		return false, nil
	}

	// Request file.
	res, err := cl.Get(url)
	if err != nil {
		return false, fmt.Errorf("Failed to request file: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Got bad response status: %s", res.Status)
	}

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

// Downloads a file iff the 'to' path does not exist. Give a client for
// logged-in sessions, or nil to start a new session. Values will be used as
// POST form values. Returns true iff file was downloaded.
func downloadIfNotExistsPost(url, to string, cl *http.Client,
	values urllib.Values) (bool, error) {
	to = expandPath(to)
	if ts := fileTimestamp(to); ts == -1 || ts < fromTimestamp {
		return false, nil
	}

	// Create output directory.
	err := mkdir(filepath.Dir(to))
	if err != nil {
		return false, fmt.Errorf("Failed to make dir: %v", err)
	}

	// Instantiate client.
	if cl == nil {
		cl = &http.Client{}
	}

	// Check if file already exists.
	if fileExists(to) && fileSize(to) != 0 {
		return false, nil
	}

	// Request file.
	res, err := cl.PostForm(url, values)
	if err != nil {
		return false, fmt.Errorf("Failed to request file: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Got bad response status: %s", res.Status)
	}

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

// A directory and a file. Surprised? So are we!
type dirFile struct {
	dir  string
	file string
}

// mkdirLock ensures that calls to os.MkDir are made sequentially.
var mkdirLock sync.Mutex

func mkdir(path string) error {
	mkdirLock.Lock()
	defer mkdirLock.Unlock()
	return os.MkdirAll(path, 0700)
}

// ----- TIMESTAMP HANDLING ---------------------------------------------------

// fromTimestamp is the minimal time from which we start downloading. Files
// older than this timestamp are ignored.
var fromTimestamp int64 = -1

// SetFromTimestamp sets the minimal time from which files will be downloaded.
// Files older that this timestamp are ignored. Format is "YYYYMMDDhhmm".
func SetFromTimestamp(s string) error {
	t := fileTimestamp(s)
	if t == -1 {
		return fmt.Errorf("bad timestamp: %q, expected %q", s, "YYYYMMDDhhmm")
	}
	fromTimestamp = t
	return nil
}

// Infers the timestamp of a file according to its name. Returns -1 if failed.
func fileTimestamp(file string) int64 {
	match := regexp.MustCompile("(\\D|^)(20\\d{10})(\\D|$)").FindStringSubmatch(filepath.Base(file))
	if match == nil || len(match[2]) != 12 {
		return -1
	}
	digits := match[2]
	year, _ := strconv.ParseInt(digits[0:4], 10, 64)
	month, _ := strconv.ParseInt(digits[4:6], 10, 64)
	day, _ := strconv.ParseInt(digits[6:8], 10, 64)
	hour, _ := strconv.ParseInt(digits[8:10], 10, 64)
	minute, _ := strconv.ParseInt(digits[10:12], 10, 64)
	t := time.Date(int(year), time.Month(month), int(day), int(hour),
		int(minute), 0, 0, time.UTC)

	return t.Unix()
}

// expandPath replaces {{date}} in a path with an actual date, inferred from its timestamp.
// If no timestamp can be inferred, returns the path as is.
func expandPath(p string) string {
	ts := fileTimestamp(p)
	date := ""
	if ts == -1 {
		date = "unknown-date"
	} else {
		t := time.Unix(ts, 0)
		date = t.Format("2006-01-02")
	}
	p = strings.Replace(p, "{{date}}", date, -1)
	return p
}

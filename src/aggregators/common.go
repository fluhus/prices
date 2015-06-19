// Deals with downloading data files from the different servers, with specific
// configurations for each chain.
package aggregators

import (
	"net/http/cookiejar"
	"net/url"
	"net/http"
	"regexp"
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
// path is malformed.
func singleCookieJar(path, name, value string) (http.CookieJar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil { return nil, err }
	pathUrl, err := url.Parse(path)
	if err != nil { return nil, err }
	jar.SetCookies(pathUrl, []*http.Cookie{&http.Cookie{
		Name: name, Value: value}})
	return jar, nil
}


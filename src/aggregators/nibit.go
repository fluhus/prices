package aggregators

// An aggregator for Nibit-based chains.

import (
	"net/http"
	//"io/ioutil"
	"fmt"
	"regexp"
	//"path/filepath"
	//"log"
	//"os"
)

const nibitHome = "http://matrixcatalog.co.il/"

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
	
	cl, err := a.startSession()
	if err != nil {
		return fmt.Errorf("Failed to start session: %v", err)
	}
	
	fmt.Println(cl.Jar)
	//fmt.Println(res.Header)

	return nil
}

// Returns a client with a session ID cookie.
func (a *nibitAggregator) startSession() (*http.Client, error) {
	// Get homepage.
	res, err := http.Head("http://matrixcatalog.co.il/NBCompetitionRegulations.aspx")
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


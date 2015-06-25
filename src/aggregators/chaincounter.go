package aggregators

// Not an aggregator. Counts chains on the authority's page, to alert when new
// chains are added.

import (
	"net/http"
	"regexp"
	"fmt"
	"io/ioutil"
)

// Page with table of chains.
const chainsPage = "http://www.economy.gov.il/Trade/ConsumerProtection/Pages/PriceTransparencyRegulations.aspx"

// Counts rows in the chain table on the authority's page.
func CountChains() (int, error) {
	// Get page.
	res, err := http.Get(chainsPage)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Got status: %s", res.Status)
	}
	
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("Failed to read body: %v", err)
	}
	
	// Parse page.
	rows := regexp.MustCompile("<tr class=\"ms-rteTable(Even|Odd)Row-mytable" +
			"\">").FindAllSubmatch(body, -1)
	
	if len(rows) == 0 {
		return 0, fmt.Errorf("Found 0 chains; page structure may have changed.")
	}
	
	return len(rows), nil
}


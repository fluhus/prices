package aggregators

// An aggregator for the Co-Op chain.

import (
	"net/http"
	"io"
	"fmt"
	"os"
)

type CoopAggregator struct{}

// Returns a new Co-Op aggregator.
func NewCoopAggregator() *CoopAggregator {
	return &CoopAggregator{}
}

func (a *CoopAggregator) Aggregate(dir string) error {
	jar := singleCookieJar("http://coopisrael.coop",
			"_gat", "GA1.2.1410125988.1434778651")
	cl := &http.Client{Jar: jar}
	// cl := &http.Client{}
	res, err := cl.Get("http://coopisrael.coop/home/get_promo/?branch=2&type=gzip&agree=1")
	if err != nil { return err }
	defer res.Body.Close()
	
	fmt.Println(res.Status)
	io.Copy(os.Stdout, res.Body)

	return nil
}

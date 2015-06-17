// Deals with downloading data files from the different servers, with specific
// configurations for each chain.
package aggregators

// An aggregator downloads data files for a specific chain.
type Aggregator interface {
	Aggregate(dir string) error
}

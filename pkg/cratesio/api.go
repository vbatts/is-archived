package cratesio

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

/*
curl -sSL "https://crates.io/api/v1/crates?per_page=100&page=1" | \
jq '.crates[] | select(.repository != null) ' | less

curl -sSL "https://crates.io/api/v1/crates/aarch64-esr-decoder" | jq . | less
*/

const apiEndpoint = "https://crates.io/api/v1"

// Crate is a _wildly_ minimal representation of the data structure returned in
// the crates.io API endpoint for listing or single package.
type Crate struct {
	ID         string `json:"id"`
	Repository string `json:"repository"`
}

// Single is when querying on a specific crate package.
type Single struct {
	Crate Crate `json:"crate"`
}

// Listing is the structure on listing all crates from their crates.io API
// endpoint.
type listing struct {
	Crates []Crate     `json:"crates"`
	Meta   listingMeta `json:"meta"`
}

// ListingMeta is metadata needed for collecting all the crates listings.
type listingMeta struct {
	NextPage string `json:"next_page"` // URL query parameters to get to the next page of results. null is the end.
	PrevPage string `json:"prev_page"` // URL query parameters to get to the previous page of results. null is the beginning.
	Total    int64  `json:"total"`     // _current_ total as of this query. _presumably_ doesn't change during a collection :-D
}

// LoadSingle populates a Single structure from the provided io.Reader
func LoadSingle(rdr io.Reader) (*Single, error) {
	s := Single{}
	buf, err := io.ReadAll(rdr)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// FetchSingle makes a call to the crates.io API regarding pkgname, and populates a returned Single.
func FetchSingle(pkgname string) (*Single, error) {
	u := fmt.Sprintf("%s/crates/%s", apiEndpoint, pkgname)
	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("fetching %q: %w", u, err)
	}
	defer resp.Body.Close()

	s, err := LoadSingle(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("loading %q: %w", u, err)
	}
	return s, nil
}

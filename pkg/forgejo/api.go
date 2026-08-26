package forgejo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/vbatts/is-archived/version"
)

// Repo is a minimal representation of the fields used from a
// forgejo/gitea/gogs repo API response, e.g.:
// https://git.batts.cloud/api/v1/repos/vbatts/too-soon
// See testdata/forgejo.json for a full example response.
type Repo struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
	Archived bool   `json:"archived"`
}

// FetchRepo queries <scheme>://<host>/api/v1/repos/<org>/<repo>, the repo-info
// endpoint shared by Forgejo, Gitea, and (for the fields used here) Gogs.
//
// ok is false whenever the response doesn't look like a forgejo/gitea/gogs
// repo document -- which is the expected outcome for the vast majority of
// hosts, since this is tried as a fallback against any host not already
// known to be something else. That is not treated as an error; a non-nil
// error is reserved for failures worth surfacing (e.g. a malformed request).
func FetchRepo(scheme, host, org, repo string) (*Repo, bool, error) {
	if scheme == "" {
		scheme = "https"
	}
	u := fmt.Sprintf("%s://%s/api/v1/repos/%s/%s", scheme, host, org, repo)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, false, fmt.Errorf("fetching %q: %w", u, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", version.Project, version.Version))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// most likely just a host that isn't running forgejo/gitea/gogs
		return nil, false, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, nil
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, nil
	}

	r := Repo{}
	if err := json.Unmarshal(buf, &r); err != nil {
		return nil, false, nil
	}
	if r.FullName == "" {
		// 200 but doesn't look like a repo document
		return nil, false, nil
	}
	return &r, true, nil
}

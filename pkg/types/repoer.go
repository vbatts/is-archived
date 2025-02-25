package types

import (
	"fmt"

	"github.com/vbatts/is-archived/pkg/check"
)

type Repoer interface {
	// Run will evaluate whether the package is archive only if the repository is hosted on that Repoer's domain i.e. github.com, codeberg.org, etc
	Run(*check.Check) error

	Domain() string
}

var repoers = []Repoer{}

func RegisterRepoer(r Repoer) {
	repoers = append(repoers, r)
}

func RepoerDomains() []string {
	domains := []string{}
	for _, r := range repoers {
		domains = append(domains, r.Domain())
	}
	return domains
}

func RepoerRun(ck *check.Check) error {
	if ck == nil || ck.VcsUrl == nil {
		return fmt.Errorf("check and/or its URL is nil")
	}
	host := ck.VcsUrl.Host
	found := false
	for _, rp := range repoers {
		if host != rp.Domain() {
			continue
		}
		found = true

		rp.Run(ck)
	}
	if !found {
		return fmt.Errorf("no checks run for %q", ck.PkgName)
	}
	return nil
}

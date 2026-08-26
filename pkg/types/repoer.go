package types

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/vbatts/is-archived/pkg/check"
)

type Repoer interface {
	// Run will evaluate whether the package is archive only if the repository is hosted on that Repoer's domain i.e. github.com, codeberg.org, etc
	Run(*check.Check) error

	// Domain returns the fixed host this Repoer handles, e.g. "github.com".
	// A Repoer that can't be pinned to one host (e.g. a self-hosted
	// forgejo/gitea/gogs instance, which could be at any domain) returns ""
	// to mark itself as a wildcard: it is tried only as a fallback, for
	// hosts no domain-specific Repoer claimed, and is excluded from
	// RepoerDomains().
	Domain() string
}

var repoers = []Repoer{}

func RegisterRepoer(r Repoer) {
	repoers = append(repoers, r)
}

// RepoerDomains lists the fixed hosts covered by registered Repoers.
// Wildcard Repoers (Domain() == "") are omitted, since they don't have one.
func RepoerDomains() []string {
	domains := []string{}
	for _, r := range repoers {
		if r.Domain() == "" {
			continue
		}
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
		if rp.Domain() == "" || rp.Domain() != host {
			continue
		}
		found = true

		if err := rp.Run(ck); err != nil {
			logrus.Warnf("when checking %q: %s", ck.PkgName, err)
		}
	}
	if found {
		return nil
	}

	// no domain-specific Repoer claimed this host; give wildcard Repoers a
	// chance to probe it (e.g. a self-hosted forgejo/gitea/gogs instance).
	for _, rp := range repoers {
		if rp.Domain() != "" {
			continue
		}
		found = true

		if err := rp.Run(ck); err != nil {
			logrus.Warnf("when checking %q: %s", ck.PkgName, err)
		}
	}
	if !found {
		return fmt.Errorf("no checks run for %q", ck.PkgName)
	}
	return nil
}

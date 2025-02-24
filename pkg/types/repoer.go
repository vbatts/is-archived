package types

import "github.com/vbatts/is-archived/pkg/check"

type Repoer interface {
	// Run will evaluate whether the package is archive only if the repository is hosted on that Repoer's domain i.e. github.com, codeberg.org, etc
	Run(*check.Check)
}

var repoers = []Repoer{}

func RegisterRepoer(r Repoer) {
	repoers = append(repoers, r)
}

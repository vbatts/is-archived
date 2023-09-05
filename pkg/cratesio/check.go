package cratesio

import (
	"fmt"
	urlpkg "net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vbatts/is-archived/pkg/check"
)

func ToCheck(c *Cargo) ([]check.Check, error) {
	toCheck := []check.Check{}

	if c.Package.Repository != "" {
		u, err := urlpkg.Parse(c.Package.Repository)
		if err != nil {
			logrus.Infof("%q did not parse correctly", c.Package.Repository)
		} else {
			toCheck = append(toCheck, check.Check{
				Lang:    c.Package.Name,
				PkgName: c.Package.Repository,
				VcsUrl:  u,
			})
		}
	}

	for dep := range c.Dependencies {
		s, err := FetchSingle(dep)
		if err != nil {
			return toCheck, fmt.Errorf("failed fetching %q: %w", dep, err)
		}
		if strings.ToLower(s.Crate.ID) != strings.ToLower(dep) {
			logrus.Infof("%q does not match %q", s.Crate.ID, dep)
			continue
		}
		if s.Crate.Repository == "" {
			logrus.Infof("%q does not list a repository", s.Crate.ID)
			continue
		}
		u, err := urlpkg.Parse(s.Crate.Repository)
		if err != nil {
			logrus.Infof("%q did not parse correctly", s.Crate.Repository)
			continue
		}
		toCheck = append(toCheck, check.Check{
			Lang:    Name,
			PkgName: dep,
			VcsUrl:  u,
		})
	}

	return toCheck, nil
}

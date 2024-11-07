package cratesio

import (
	"fmt"
	urlpkg "net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vbatts/is-archived/pkg/check"
)

func ToCheckCargo(c *Cargo) ([]check.Check, error) {
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
		if !strings.EqualFold(s.Crate.ID, dep) {
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

func ToCheckCargoLock(cl *CargoLock) ([]check.Check, error) {
	toCheck := []check.Check{}

	for _, pkg := range cl.Package {
		s, err := FetchSingle(pkg.Name)
		if err != nil {
			return toCheck, fmt.Errorf("failed fetching %q: %w", pkg, err)
		}
		if !strings.EqualFold(s.Crate.ID, pkg.Name) {
			logrus.Infof("%q does not match %q", s.Crate.ID, pkg.Name)
			continue
		}
		if s.Crate.Repository == "" {
			logrus.Infof("%q does not list a repository", s.Crate.ID)
			continue
		}
		/* was beginning on this path, but they can have a number of schema/protocols, that are not https/http
		if pkg.Source != "" && !pkg.IsSourceRegistry() {
			u, err := urlpkg.Parse(pkg.Source)
		}
		*/
		u, err := urlpkg.Parse(s.Crate.Repository)
		if err != nil {
			logrus.Infof("%q did not parse correctly", s.Crate.Repository)
			continue
		}
		if strings.HasSuffix(u.Path, ".git") {
			i := strings.LastIndex(u.Path, ".git")
			u.Path = u.Path[:i]
		}
		toCheck = append(toCheck, check.Check{
			Lang:    Name,
			PkgName: pkg.Name,
			VcsUrl:  u,
		})
	}

	return toCheck, nil
}

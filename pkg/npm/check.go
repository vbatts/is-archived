package npm

import (
	urlpkg "net/url"

	"github.com/sirupsen/logrus"
	"github.com/vbatts/is-archived/pkg/check"
)

// ToCheckNpm collects all the dependencies to check from the provided package.
// This does not (yet) recursively check all deps of deps, only the first layer of deps.
// This includes `dependencies
func ToCheckNpm(p *Package) ([]check.Check, error) {
	toCheck := []check.Check{}

	if p.Repository.URL != "" {
		u, err := urlpkg.Parse(p.Repository.URL)
		if err != nil {
			logrus.Infof("%q did not parse correctly", p.Repository.URL)
		} else {
			toCheck = append(toCheck, check.Check{
				Lang:    Name,
				PkgName: p.Name,
				VcsUrl:  u,
			})
		}
	}

	allOfThem := keysFromMaps(p.Dependencies, p.DevDependencies, p.OptionalDependencies)
	logrus.Infof("  fetching metadata on %d packages", len(allOfThem))
	for _, dep := range allOfThem {
		s, err := FetchSingle(dep)
		if err != nil {
			logrus.Errorf("failed fetching %q: %s", dep, err)
			continue
		}
		if s.Repository.URL != "" {
			u, err := urlpkg.Parse(s.Repository.URL)
			if err != nil {
				logrus.Infof("%q did not parse correctly", s.Repository.URL)
			} else {
				toCheck = append(toCheck, check.Check{
					Lang:    Name,
					PkgName: p.Name,
					VcsUrl:  u,
				})
			}
		}
	}
	return toCheck, nil
}

func keysFromMaps(sets ...map[string]string) []string {
	names := []string{}
	for _, set := range sets {
		for key := range set {
			names = append(names, key)
		}
	}
	return names
}

package cratesio

import (
	"encoding/json"
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
				Lang:    Name,
				PkgName: c.Package.Name,
				VcsUrl:  u,
			})
		}
	}

	// holding bin for all the various Dependencies
	deps := map[string]interface{}{}

	// Collect all the Cargo Dependencies before cleaning them up
	if l := len(c.Dependencies); l > 0 {
		logrus.Debugf("reviewing %d base deps", l)
	}
	for dep, val := range c.Dependencies {
		// we can set it directly here because it's empty and should not have
		// duplicates in the base set
		deps[dep] = val
	}
	if l := len(c.BuildDependencies); l > 0 {
		logrus.Debugf("reviewing %d build-dependencies", l)
	}
	for dep, val := range c.BuildDependencies {
		deps[dep] = val
	}
	if l := len(c.DevDependencies); l > 0 {
		logrus.Debugf("reviewing %d dev-dependencies", l)
	}
	for dep, val := range c.DevDependencies {
		deps[dep] = val
	}
	if l := len(c.Workspace.Dependencies); l > 0 {
		logrus.Debugf("reviewing %d workspace deps", l)
	}
	for dep, val := range c.Workspace.Dependencies {
		deps[dep] = val
	}
	for target, t := range c.Target {
		logrus.Debugf("reviewing deps of %s: %#v", target, t)
		for dep, val := range t.Dependencies {
			deps[dep] = val
		}
	}

	for dep, val := range deps {
		dd, err := getDepData(val)
		if err != nil {
			logrus.Infof("%q value was not parsed correctly (%#v): %s", dep, val, err)
			continue
		}
		logrus.Debugf("%q value: %#v", dep, dd)
		s, err := FetchSingle(dep)
		if err != nil {
			logrus.Warnf("failed fetching %q: %s", dep, err)
			continue
		}
		if !strings.EqualFold(s.Crate.ID, dep) {
			logrus.Infof("%q does not appear to be on crates.io ... skipping.", dep)
			continue
		}
		if s.Crate.Repository == "" {
			logrus.Infof("%q does not list a repository on %s/%s", s.Crate.ID, baseUrl, s.Crate.ID)
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

	toCheck = uniqueChecks(toCheck)

	return toCheck, nil
}

func ToCheckCargoLock(cl *CargoLock) ([]check.Check, error) {
	toCheck := []check.Check{}

	for _, pkg := range cl.Package {
		s, err := FetchSingle(pkg.Name)
		if err != nil {
			logrus.Warnf("failed fetching %q: %s", pkg.Name, err)
			continue
		}

		var u *urlpkg.URL
		switch {
		case strings.EqualFold(s.Crate.ID, pkg.Name) && s.Crate.Repository != "":
			// found on crates.io, and it lists a repository
			u, err = urlpkg.Parse(s.Crate.Repository)
			if err != nil {
				logrus.Infof("%q did not parse correctly", s.Crate.Repository)
				continue
			}
		case pkg.IsGitHttps() || pkg.IsRegistryHttps():
			// not resolvable through crates.io metadata, but the lockfile's
			// own `source` (e.g. "git+https://..." or "registry+https://...")
			// points at a fetchable location
			src := pkg.Source
			if i := strings.Index(src, "+"); i >= 0 {
				src = src[i+1:]
			}
			u, err = urlpkg.Parse(src)
			if err != nil {
				logrus.Infof("%q did not parse correctly", pkg.Source)
				continue
			}
		default:
			if !strings.EqualFold(s.Crate.ID, pkg.Name) {
				logrus.Infof("%q does not appear to be on crates.io ... skipping.", pkg.Name)
			} else {
				logrus.Infof("%q does not list a repository on %s/%s", s.Crate.ID, baseUrl, s.Crate.ID)
			}
			continue
		}

		if strings.HasSuffix(u.Path, ".git") {
			u.Path = strings.TrimSuffix(u.Path, ".git")
		}
		toCheck = append(toCheck, check.Check{
			Lang:    Name,
			PkgName: pkg.Name,
			VcsUrl:  u,
		})
	}

	toCheck = uniqueChecks(toCheck)

	return toCheck, nil
}

func uniqueChecks(toCheck []check.Check) []check.Check {
	m := map[string]check.Check{}
	for _, ck := range toCheck {
		m[ck.PkgName] = ck
	}
	cks := []check.Check{}
	for _, v := range m {
		cks = append(cks, v)
	}
	return cks
}

/*
this is a bit gross due to the flexibility of toml ...
```toml
[dependencies]
hello_utils = "0.1.0"
smallvec = { git = "https://github.com/servo/rust-smallvec.git", version = "1.0" }
bitflags = { path = "my-bitflags", version = "1.0" }
```
*/
func getDepData(d interface{}) (*depData, error) {
	// "1.2.0"
	s, ok := d.(string)
	if ok {
		return &depData{Version: s}, nil
	}
	// 1.0
	f, ok := d.(float64)
	if ok {
		return &depData{Version: fmt.Sprintf("%f", f)}, nil
	}
	// 1
	i, ok := d.(int64)
	if ok {
		return &depData{Version: fmt.Sprintf("%d", i)}, nil
	}

	buf, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	dd := depData{}
	err = json.Unmarshal(buf, &dd)
	if err != nil {
		return nil, err
	}
	return &dd, nil
}

type depData struct {
	Version  string   `json:"version,omitempty"`
	Path     string   `json:"path,omitempty"`
	Git      string   `json:"git,omitempty"`
	Optional bool     `json:"optional,omitempty"`
	Registry string   `json:"registry,omitempty"`
	Package  string   `json:"package,omitempty"`
	Features []string `json:"features,omitempty"`
}

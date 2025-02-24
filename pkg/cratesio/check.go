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

	for dep, val := range c.Dependencies {
		dd, err := getDepData(val)
		if err != nil {
			logrus.Infof("%q value was not parsed correctly (%#v): %s", dep, val, err)
			continue
		}
		logrus.Debugf("%q value: %#v", dep, dd)
		s, err := FetchSingle(dep)
		if err != nil {
			return toCheck, fmt.Errorf("failed fetching %q: %w", dep, err)
		}
		if !strings.EqualFold(s.Crate.ID, dep) {
			logrus.Infof("%q does not match %q. Skipping.", s.Crate.ID, dep)
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

	for target, t := range c.Target {
		logrus.Debugf("reviewing deps of %s: %#v", target, t)
		for dep, val := range t.Dependencies {
			dd, err := getDepData(dep)
			if err != nil {
				logrus.Infof("%q value was not parsed correctly (%#v): %s", dep, val, err)
				continue
			}
			logrus.Debugf("%q value: %#v", dep, dd)
			s, err := FetchSingle(dep)
			if err != nil {
				return toCheck, fmt.Errorf("failed fetching %q: %w", dep, err)
			}
			if !strings.EqualFold(s.Crate.ID, dep) {
				logrus.Infof("%q does not match %q. Skipping.", s.Crate.ID, dep)
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
			logrus.Infof("%q does not match %q. Skipping.", s.Crate.ID, pkg.Name)
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

package cratesio

import (
	"fmt"
	urlpkg "net/url"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vbatts/is-archived/pkg/check"
)

func init() {
	if os.Getenv("DEBUG") != "" {
		logrus.SetOutput(os.Stderr)
		logrus.SetLevel(logrus.DebugLevel)
	}
}

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

	for dep, val := range c.Dependencies {
		dd, err := getDepData(val)
		if err != nil {
			logrus.Infof("%q value was not parsed correctly: %s", dep, err)
			continue
		}
		logrus.Debugf("%q value: %#v", dep, dd)
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

	for target, deps := range c.Target {
		logrus.Debugf("reviewing deps of %s: %#v", target, deps)
		//for dep := range deps {
		//}
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

	m, ok := d.(map[string]string)
	if ok {
		dd := depData{}
		for k, v := range m {
			switch {
			case strings.ToLower(k) == "version":
				dd.Version = v
			case strings.ToLower(k) == "git":
				dd.Git = v
			case strings.ToLower(k) == "path":
				dd.Path = v
			case strings.ToLower(k) == "registry":
				dd.Registry = v
			case strings.ToLower(k) == "package":
				dd.Package = v
			}
		}
		return &dd, nil
	}
	return nil, fmt.Errorf("unknown dependency value")
}

type depData struct {
	Version  string
	Path     string
	Git      string
	Optional bool
	Registry string
	Package  string
}

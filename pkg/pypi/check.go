package pypi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vbatts/is-archived/pkg/check"
	"github.com/vbatts/is-archived/pkg/types"
	"github.com/vbatts/is-archived/version"
)

// https://docs.pypi.org/api/json/

const (
	Name        = "Python pip/pypi"
	PythonFile  = "requirements.txt"
	apiEndpoint = "https://pypi.org/pypi"
)

func init() {
	if err := types.RegisterPackager(pythonPackager{fileType: PythonFile}); err != nil {
		logrus.Errorf("failed to register Packager for %q", PythonFile)
	}
}

type pythonPackager struct {
	fileType string
}

func (pp pythonPackager) Name() string {
	return Name
}

func (pp pythonPackager) FileType() string {
	return pp.fileType
}

func (pp pythonPackager) LoadFile(filename string) ([]check.Check, error) {
	pf, err := LoadRequirementsFile(filename)
	if err != nil {
		return nil, err
	}
	return ToCheck(pf)
}

func LoadRequirementsFile(filename string) ([]string, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	pkgs := []string{}
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		txt := scanner.Text()
		if txt == "" {
			continue
		}
		pkgs = append(pkgs, parseReqLine(txt))
	}
	return pkgs, nil
}

var parseTokens = []string{";", "<", "<=", ">", ">=", "="}

func parseReqLine(line string) string {
	// TODO this needs splitting on ';' and equalit symbols
	logrus.Debugf("[pypi] before: %q", line)
	for _, tok := range parseTokens {
		if strings.Contains(line, tok) {
			line = strings.Split(line, tok)[0]
		}
	}
	logrus.Debugf("[pypi] after: %q", line)
	return line
}

func ToCheck(reqs []string) ([]check.Check, error) {
	checks := []check.Check{}
	for _, req := range reqs {
		pkg, err := FetchSingle(req)
		if err != nil {
			logrus.Errorf("[pypi] failed to fetch for %q", req)
			continue
		}
		confidence := 0.0
		check := check.Check{
			Lang:    Name,
			PkgName: pkg.Info.Name,
			//VcsUrl:  u,
		}
		if pkg.Info.HomePage != "" {
			logrus.Debugf("[pypi] %q - Homepage: %q", req, pkg.Info.HomePage)
			u, err := urlpkg.Parse(pkg.Info.HomePage)
			if err != nil {
				logrus.Errorf("[pypi] %q Error parsing URL %q", req, pkg.Info.HomePage)
			} else {
				// TODO iterate through the domains of the Repoers
				if u.Host == "github.com" {
					confidence = 1.0
					check.VcsUrl = u
				}
			}
		}
		for k, v := range pkg.Info.ProjectURLs {
			// TODO get repo from pkg.Info.ProjectURLs where key is like "Github: repo" ??
			// which is sloppy. These keys seem like openended strings:
			// "source", "github", "Github: repo", etc
			logrus.Debugf("[pypi] %q - found projectURL %q", req, k)
			if confidence < 1.0 {
				for _, pun := range projectUrlNames {
					if strings.EqualFold(pun, k) && v != "" {
						logrus.Debugf("[pypi] %q - trying %q : %s", req, k, v)
						u, err := urlpkg.Parse(v)
						if err != nil {
							logrus.Errorf("[pypi] %q Error parsing URL %q", req, v)
						} else {
							// TODO iterate through the domains of the Repoers
							if u.Host == "github.com" {
								check.VcsUrl = u
							}
						}
					}
				}
			}
		}
		if check.VcsUrl == nil {
			logrus.Infof("  could not determine source repo for %q", check.PkgName)
			continue
		}
		checks = append(checks, check)
	}
	return checks, nil
}

var projectUrlNames = []string{"source", "github", "Github: repo", "homepage"}

type Package struct {
	Info Info `json:"info"`
}

type Info struct {
	Name            string            `json:"name"`
	Summary         string            `json:"summary"`
	RequiresDist    []string          `json:"requires_dist"`
	HomePage        string            `json:"home_page"`
	Maintainer      string            `json:"maintainer"`
	MaintainerEmail string            `json:"maintainer_email"`
	ProjectURLs     map[string]string `json:"project_urls"`

	// ... obviously load more
}

func FetchSingle(pkg string) (*Package, error) {
	u := fmt.Sprintf("%s/%s/json", apiEndpoint, pkg)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching %q: %w", u, err)
	}
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", version.Project, version.Version))
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//logrus.Infof("%#v", string(buf))

	p := Package{}
	err = json.Unmarshal(buf, &p)
	if err != nil {
		return nil, err
	}

	//logrus.Info(p.Repository.URL)
	return &p, nil
}

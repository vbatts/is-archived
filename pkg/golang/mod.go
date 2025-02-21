package golang

import (
	"encoding/json"
	"fmt"
	urlpkg "net/url"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vbatts/is-archived/pkg/check"
	"github.com/vbatts/is-archived/pkg/types"
	"github.com/vbatts/is-archived/pkg/vcs"
)

const (
	Name      = "go.mod (golang)"
	GoModFile = "go.mod"
)

func init() {
	if err := types.RegisterPackager(goPackager{fileType: GoModFile}); err != nil {
		logrus.Errorf("failed to register Packager for %q", GoModFile)
	}
}

type goPackager struct {
	fileType string
}

func (gp goPackager) Name() string {
	return Name
}

func (gp goPackager) FileType() string {
	return gp.fileType
}

func (gp goPackager) LoadFile(filename string) ([]check.Check, error) {
	gmf, err := LoadGoModFile(filename)
	if err != nil {
		return nil, err
	}
	return ToCheck(gmf)
}

func LoadGoModFile(fpath string) (*Mod, error) {
	cmd := exec.Command("go", "mod", "edit", "-json")
	base := filepath.Base(fpath)
	if base != "." && base != "go.mod" {
		// then we're not running against our current directory
		cmd.Dir = fpath // maybe should insure this is a directory :-D
	}

	buf, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed getting go.mod from %q: %w", fpath, err)
	}
	// get modfile struct
	m := Mod{}
	err = json.Unmarshal(buf, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %q: %w", fpath, err)
	}
	return &m, nil
}

type Mod struct {
	Require []Import `json:"Require"`
}

type Import struct {
	Path     string `json:"Path"`
	Version  string `json:"Version"`
	Indirect bool   `json:"Indirect"`
}

func ToCheck(m *Mod) ([]check.Check, error) {
	// collect the list first
	toCheck := []check.Check{}
	for _, req := range m.Require {
		// TODO this ought to iterate over Repoer's domains as the "known" domains to check.
		if !strings.HasPrefix(req.Path, "github.com") {
			_, mi, err := vcs.MetaImportForPath(req.Path)
			if err == nil { // ignoring this error as we'll just skip and continue ...
				if len(mi) == 0 {
					// we didn't get any meta imports from that import path
					logrus.Debugf("skipping %q as it had no HTML Meta go-imports", req.Path)
					continue
				}
				u, err := urlpkg.Parse(mi[0].RepoRoot)
				if err != nil {
					logrus.Infof("skipping %q as %q didn't parse well: %v", req.Path, mi[0].RepoRoot, err)
					continue
				}
				toCheck = append(toCheck, check.Check{
					Lang:    Name,
					PkgName: req.Path,
					VcsUrl:  u,
				})
			}
			continue
		}

		u, err := urlpkg.Parse(fmt.Sprintf("https://%s", req.Path))
		if err != nil {
			logrus.Debugf("skipping %q as %q didn't parse well: %v", req.Path, fmt.Sprintf("https://%s", req.Path), err)
			continue
		}
		toCheck = append(toCheck, check.Check{
			Lang:    Name,
			PkgName: req.Path,
			VcsUrl:  u,
		})
	}
	return toCheck, nil
}

package npm

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/vbatts/is-archived/pkg/check"
	"github.com/vbatts/is-archived/pkg/types"
)

// https://docs.npmjs.com/cli/v10/configuring-npm/package-json?v=true

const (
	Name           = "npm package.json (nodejs)"
	NpmPackageFile = "package.json"
)

func init() {
	if err := types.RegisterPackager(npmPackager{fileType: NpmPackageFile}); err != nil {
		logrus.Errorf("failed to register Packager for %q", NpmPackageFile)
	}
}

type npmPackager struct {
	fileType string
}

func (np npmPackager) Name() string {
	return Name
}

func (np npmPackager) FileType() string {
	return np.fileType
}

func (np npmPackager) LoadFile(filename string) ([]check.Check, error) {
	pf, err := LoadPackageJSON(filename)
	if err != nil {
		return nil, err
	}
	return ToCheckNpm(pf)
}

func LoadPackageJSON(filename string) (*Package, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	p := Package{}
	err = json.Unmarshal(buf, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

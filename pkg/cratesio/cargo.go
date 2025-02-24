package cratesio

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"github.com/vbatts/is-archived/pkg/check"
	"github.com/vbatts/is-archived/pkg/types"
)

// https://doc.rust-lang.org/cargo/reference/manifest.html
// https://doc.rust-lang.org/cargo/reference/specifying-dependencies.html

const (
	// Name for identifying this language support
	Name = "crates.io (Rust)"

	CargoFile     = "Cargo.toml"
	CargoLockFile = "Cargo.lock"
)

func init() {
	if err := types.RegisterPackager(cargoPackager{fileType: CargoFile}); err != nil {
		logrus.Errorf("failed to register Packager for %q", CargoFile)
	}
	if err := types.RegisterPackager(cargoPackager{fileType: CargoLockFile}); err != nil {
		logrus.Errorf("failed to register Packager for %q", CargoLockFile)
	}
}

type cargoPackager struct {
	fileType string
}

func (cp cargoPackager) Name() string {
	return fmt.Sprintf("%s - %s", Name, cp.fileType)
}

func (cp cargoPackager) FileType() string {
	return cp.fileType
}

func (cp cargoPackager) LoadFile(filename string) ([]check.Check, error) {
	if cp.fileType == CargoLockFile {
		clf, err := LoadCargoLockFile(filename)
		if err != nil {
			return nil, err
		}
		return ToCheckCargoLock(clf)
	}

	cf, err := LoadCargoFile(filename)
	if err != nil {
		return nil, err
	}
	return ToCheckCargo(cf)
}

// Cargo is a representation of a `Cargo.toml`.
// This is bare-bones enough to gather the names of the dependencies
type Cargo struct {
	Package           Package                `toml:"package"`
	Dependencies      map[string]interface{} `toml:"dependencies"`
	BuildDependencies map[string]interface{} `toml:"dev-dependencies"`
	DevDependencies   map[string]interface{} `toml:"build-dependencies"`
	Target            map[string]Target      `toml:"target"`
}

type Target struct {
	Dependencies map[string]interface{} `toml:"dependencies"`
}

// Package is a bare couple of fields from a `Cargo.toml`
type Package struct {
	Name       string `toml:"name"`
	Version    string `toml:"version"`
	Edition    string `toml:"edition"`
	Repository string `toml:"repository"`
	Source     string `toml:"source,omitempty"`
	Checksum   string `toml:"checksum,omitempty"`
}

// IsSourceRegistry checks whether the Source field is referring
// to a cargo registry index, or to a specific repo.
func (p *Package) IsSourceRegistry() bool {
	return strings.HasPrefix(p.Source, "registry+http")
}

type CargoLock struct {
	Version int64     `toml:"version"`
	Package []Package `toml:"package"`
}

// LoadCargoFile reads filename (usually "Cargo.toml") and populates the returned Cargo
func LoadCargoFile(filename string) (*Cargo, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	return LoadCargoToml(fh)
}

// LoadCargoLockFile reads filename (usually "Cargo.lock") and populates the returned CargoLock
func LoadCargoLockFile(filename string) (*CargoLock, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	return LoadCargoLock(fh)
}

// LoadCargoToml populates a Cargo structure from an io.Reader of the "Cargo.toml" type file
func LoadCargoToml(rdr io.Reader) (*Cargo, error) {
	buf, err := io.ReadAll(rdr)
	if err != nil {
		return nil, err
	}
	c := Cargo{}
	_, err = toml.Decode(string(buf), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// LoadCargoLock populates a Cargo structure from an io.Reader of the "Cargo.lock" type file
func LoadCargoLock(rdr io.Reader) (*CargoLock, error) {
	buf, err := io.ReadAll(rdr)
	if err != nil {
		return nil, err
	}

	c := CargoLock{}
	_, err = toml.Decode(string(buf), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

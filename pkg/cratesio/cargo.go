package cratesio

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"
)

// Name for identifying this language support
const Name = "crates.io (Rust)"

// Cargo is a representation of a `Cargo.toml`.
// This is bare-bones enough to gather the names of the dependencies
type Cargo struct {
	Package      Package                `toml:"package"`
	Dependencies map[string]interface{} `toml:"dependencies"`
}

// Package is a bare couple of fields from a `Cargo.toml`
type Package struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
	Edition string `toml:"edition"`
}

// LoadCargoFile reads filename and populates the returned Cargo
func LoadCargoFile(filename string) (*Cargo, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	return LoadCargo(fh)
}

// LoadCargo reads from the io.Reader and populates the returned Cargo
func LoadCargo(rdr io.Reader) (*Cargo, error) {
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

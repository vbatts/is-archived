package types

import (
	"fmt"

	"github.com/vbatts/is-archived/pkg/check"
)

type Packager interface {
	Name() string
	FileType() string
	LoadFile(string) ([]check.Check, error)
}

var packagers = map[string]Packager{}

func RegisterPackager(p Packager) error {
	if p.FileType() == "" {
		return fmt.Errorf("no FileType")
	}
	packagers[p.FileType()] = p

	return nil
}

func PackagerFileTypes() []string {
	ft := []string{}
	for key := range packagers {
		ft = append(ft, key)
	}
	return ft
}

func GetPackager(filetype string) (Packager, error) {
	if val, ok := packagers[filetype]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("no packager for %q found", filetype)
}

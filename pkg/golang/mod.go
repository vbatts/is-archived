package golang

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
)

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

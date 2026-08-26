package cratesio

import (
	"net/url"
	"testing"
)

func TestCargoLoad(t *testing.T) {
	fpath := "testdata/Cargo.toml"
	c, err := LoadCargoFile(fpath)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]int{
		"yall":        0,
		"tokio":       0,
		"hard-xml":    0,
		"uuid":        0,
		"examplename": 0,
		"regex":       0,
	}
	// XXX is this not even working?
	for k := range c.Dependencies {
		_, ok := expected[k]
		if !ok {
			t.Errorf("expected to find %q dependency", k)
		}
	}

	if got, exp := len(c.Dependencies), 6; got != exp {
		t.Errorf("expected %d deps; got %d", exp, got)
	}
}

func TestCargoLoadWorkspace(t *testing.T) {
	fpath := "testdata/Cargo.toml.workspace"
	c, err := LoadCargoFile(fpath)
	if err != nil {
		t.Fatal(err)
	}

	if got, exp := len(c.Workspace.Dependencies), 49; got != exp {
		t.Errorf("expected %d deps; got %d", exp, got)
		t.Fatalf("%#v", c.Workspace)
	}
}

/*
From https://doc.rust-lang.org/cargo/reference/pkgid-spec.html

registry+https://github.com/rust-lang/crates.io-index#regex@1.4.3
https://github.com/rust-lang/cargo#0.52.0
https://github.com/rust-lang/cargo#cargo-platform@0.1.2
ssh://git@github.com/rust-lang/regex.git#regex@1.4.3
git+ssh://git@github.com/rust-lang/regex.git#regex@1.4.3
git+ssh://git@github.com/rust-lang/regex.git?branch=dev#regex@1.4.3
*/

func TestCargoLock(t *testing.T) {
	fpath := "testdata/Cargo.lock"
	c, err := LoadCargoLockFile(fpath)
	if err != nil {
		t.Fatalf("Failed to load Cargo.lock: %s", err)
	}
	if got, exp := len(c.Package), 744; got != exp {
		t.Errorf("expected %d packages in Cargo.lock, got %d", exp, got)
	}
	for _, pkg := range c.Package {
		if pkg.Source != "" {
			u, err := url.Parse(pkg.Source)
			if err != nil {
				t.Errorf("package source failed to parse as URL: %q", pkg.Source)
			}
			t.Logf("%s://%s%s", u.Scheme, u.Host, u.Path)
		}
	}
}

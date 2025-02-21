package cratesio

import (
	"encoding/json"
	"os"
	"testing"
)

func TestParseFetchSingleCrate(t *testing.T) {
	expected := "aarch64-esr-decoder"
	s, err := FetchSingle(expected)
	if err != nil {
		t.Fatalf("failed to fetch %q: %v", expected, err)
	}
	if s.Crate.ID != expected {
		t.Errorf("expected %q, got %q", expected, s.Crate.ID)
	}
	if s.Crate.Repository == "" {
		t.Errorf("expected %q to provide a repository address, but it is empty", expected)
	}
}

func TestParseSingleCrate(t *testing.T) {
	fpath := "testdata/aarch64-esr-decoder.crate.json"
	fh, err := os.Open(fpath)
	if err != nil {
		t.Fatalf("failed to open file %q: %v", fpath, err)
	}
	defer fh.Close()

	s, err := loadSingle(fh)
	if err != nil {
		t.Fatalf("failed to load %q: %v", fpath, err)
	}
	expected := "aarch64-esr-decoder"
	if s.Crate.ID != expected {
		t.Errorf("expected %q, got %q", expected, s.Crate.ID)
	}
	if s.Crate.Repository == "" {
		t.Errorf("expected %q to provide a repository address, but it is empty", expected)
	}
}

func TestParseListCrates(t *testing.T) {
	// we may not even _need_ to do a listing...
	// because if we load local, then we know the name of the single crate to load.
	fpath := "testdata/page1.crates.json"

	l := listing{}
	buf, err := os.ReadFile(fpath)
	if err != nil {
		t.Fatalf("failed to read file %q: %v", fpath, err)
	}
	err = json.Unmarshal(buf, &l)
	if err != nil {
		t.Fatalf("failed to unmarshal %q: %v", fpath, err)
	}

	if len(l.Crates) != 100 {
		t.Errorf("expected 100 deps, but got %d deps", len(l.Crates))
	}
	found := false
	expected := "aarch64-esr-decoder"
	for _, crate := range l.Crates {
		if crate.ID == expected {
			found = true
			if crate.Repository == "" {
				t.Errorf("expected %q to provide a repository address, but it is empty", expected)
			}
		}
	}
	if !found {
		t.Errorf("expected to find %q, but did not", expected)
	}
}

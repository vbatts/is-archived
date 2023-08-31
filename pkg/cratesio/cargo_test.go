package cratesio

import (
	"testing"
)

func TestCargoLoad(t *testing.T) {
	fpath := "testdata/Cargo.toml"
	c, err := LoadCargoFile(fpath)
	if err != nil {
		t.Fatal(err)
	}

	toCheck := map[string]int{
		"yall":        0,
		"tokio":       0,
		"hard-xml":    0,
		"uuid":        0,
		"examplename": 0,
	}
	for k := range c.Dependencies {
		_, ok := toCheck[k]
		if !ok {
			t.Errorf("expected to find %q dependency", k)
		}
	}
}

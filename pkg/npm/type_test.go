package npm

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLoadPackage(t *testing.T) {
	buf, err := os.ReadFile("./testdata/package.json")
	if err != nil {
		t.Fatal(err)
	}
	p := Package{}
	err = json.Unmarshal(buf, &p)
	if err != nil {
		t.Fatal(err)
	}

	if len(p.Dependencies) != 5 {
		t.Errorf("expected 5 dependencies, got %d", len(p.Dependencies))
	}
	//logrus.Infof("%#v", p)
}

package forgejo

import (
	"encoding/json"
	"net/url"
	"os"
	"testing"
)

func TestParseRepo(t *testing.T) {
	buf, err := os.ReadFile("testdata/forgejo.json")
	if err != nil {
		t.Fatal(err)
	}
	r := Repo{}
	if err := json.Unmarshal(buf, &r); err != nil {
		t.Fatal(err)
	}
	if expected := "vbatts/too-soon"; r.FullName != expected {
		t.Errorf("expected full_name %q, got %q", expected, r.FullName)
	}
	if r.Archived {
		t.Errorf("expected archived to be false")
	}
}

func TestOrgRepoFromURL(t *testing.T) {
	tests := []struct {
		url       string
		org, repo string
	}{
		{"https://git.batts.cloud/vbatts/too-soon", "vbatts", "too-soon"},
		{"https://git.batts.cloud/vbatts/too-soon.git", "vbatts", "too-soon"},
		{"https://git.batts.cloud/vbatts", "", ""},
	}
	for _, tt := range tests {
		u, err := url.Parse(tt.url)
		if err != nil {
			t.Fatal(err)
		}
		org, repo := orgRepoFromURL(u)
		if org != tt.org || repo != tt.repo {
			t.Errorf("%q: expected org=%q repo=%q; got org=%q repo=%q", tt.url, tt.org, tt.repo, org, repo)
		}
	}
}

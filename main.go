package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	urlpkg "net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v54/github"
	"github.com/sirupsen/logrus"
	"github.com/vbatts/is-archived/pkg/vcs"
	"golang.org/x/oauth2"
)

func init() {
	flag.Parse()
	logrus.SetOutput(os.Stderr)
}

/*
parse a `go.mod` and query github whether any of the github projects are archived.

Testing using `go mod edit -json > mod.json`.
This could be done as a file, or subshell or on std-in.
*/
func main() {
	ctx := context.Background()

	var buf []byte
	var err error

	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		logrus.Info("reading from stdin ...")
		buf, err = io.ReadAll(os.Stdin)
		if err != nil {
			logrus.Fatal(err)
		}
	} else if _, err = os.Stat("go.mod"); err == nil {
		logrus.Info("found 'go.mod'. Running `go mod edit -json'")
		cmd := exec.Command("go", "mod", "edit", "-json")
		buf, err = cmd.Output()
		if err != nil {
			logrus.Fatal(err)
		}
	} else {
		logrus.Fatal("no input provided")
	}

	// get modfile struct
	m := Mod{}
	err = json.Unmarshal(buf, &m)
	if err != nil {
		logrus.Fatal(err)
	}

	var client *github.Client
	if os.Getenv("GITHUB_TOKEN") != "" {
		// query Github for each repo
		// needs PAT for rate limiting ...

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	// collect the list first
	toCheck := []Check{}
	for _, req := range m.Require {
		if !strings.HasPrefix(req.Path, "github.com") {
			_, mi, err := vcs.MetaImportForPath(req.Path)
			if err == nil { // ignoring this error as we'll just skip and continue ...
				if len(mi) == 0 {
					// we didn't get any meta imports from that import path
					logrus.Debugf("skipping %q as it had no HTML Meta go-imports", req.Path)
					continue
				}
				u, err := urlpkg.Parse(mi[0].RepoRoot)
				if err != nil {
					logrus.Debugf("skipping %q as %q didn't parse well: %v", req.Path, mi[0].RepoRoot, err)
					continue
				}
				if u.Host != "github.com" {
					logrus.Debugf("skipping %q for now, as it isn't github (%s)", req.Path, u.Host)
					continue
				}
				toCheck = append(toCheck, Check{
					Import: req.Path,
					VcsUrl: u,
				})
			}
			continue
		}

		u, err := urlpkg.Parse(fmt.Sprintf("https://%s", req.Path))
		if err != nil {
			logrus.Debugf("skipping %q as %q didn't parse well: %v", req.Path, fmt.Sprintf("https://%s", req.Path), err)
			continue
		}
		toCheck = append(toCheck, Check{
			Import: req.Path,
			VcsUrl: u,
		})

	}

	logrus.Infof("checking %d github projects ...", len(toCheck))
	for _, check := range toCheck {
		spl := strings.Split(check.VcsUrl.Path, "/")

		repo, _, err := client.Repositories.Get(ctx, spl[1], spl[2])
		if err != nil {
			if _, ok := err.(*github.RateLimitError); ok {
				logrus.Fatal("rate limited. Try using a Personal Access Token and setting GITHUB_TOKEN env variable: ", err)
			}
			if _, ok := err.(*github.AbuseRateLimitError); ok {
				logrus.Fatal("rate limited. Try using a Personal Access Token and setting GITHUB_TOKEN env variable: ", err)
			}
			logrus.Fatal(err)
		}
		if repo.GetArchived() {
			check.Archived = true
			fmt.Printf("%q is archived (%s)\n", check.Import, check.VcsUrl.String())
		}
	}
}

type Check struct {
	Import   string
	VcsUrl   *urlpkg.URL
	Archived bool
}

type Mod struct {
	Require []Import `json:"Require"`
}

type Import struct {
	Path     string `json:"Path"`
	Version  string `json:"Version"`
	Indirect bool   `json:"Indirect"`
}

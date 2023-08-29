package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v54/github"
	"github.com/sirupsen/logrus"
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

	// get a pretty number to show
	n := 0
	for _, req := range m.Require {
		if !strings.HasPrefix(req.Path, "github.com") {
			continue
		}
		n++
	}
	logrus.Infof("checking %d github projects ...", n)

	for _, req := range m.Require {
		if !strings.HasPrefix(req.Path, "github.com") {
			continue
		}

		spl := strings.Split(req.Path, "/")

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
			if false {
				buf, err := json.MarshalIndent(repo, "", "  ")
				if err != nil {
					logrus.Fatal(err)
				}
				fmt.Println(string(buf))
			}
			fmt.Printf("%q is archived\n", req.Path)
		}
	}
}

type Mod struct {
	Require []Import `json:"Require"`
}

type Import struct {
	Path     string `json:"Path"`
	Version  string `json:"Version"`
	Indirect bool   `json:"Indirect"`
}

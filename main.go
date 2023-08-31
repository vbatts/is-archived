package main

import (
	"context"
	"flag"
	"fmt"
	urlpkg "net/url"
	"os"
	"strings"

	"github.com/google/go-github/v54/github"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"github.com/vbatts/is-archived/pkg/gh"
	"github.com/vbatts/is-archived/pkg/golang"
	"github.com/vbatts/is-archived/pkg/vcs"
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
	app := cli.App{
		Name:   "is-archived",
		Usage:  "check your project's dep's upstreams for being archived projects",
		Action: mainFunc,
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func mainFunc(c *cli.Context) error {
	ctx := context.Background()

	/* Maybe make 'go' and 'rust' subcommands to use this
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		logrus.Info("reading from stdin ...")
		buf, err = io.ReadAll(os.Stdin)
		if err != nil {
			logrus.Fatal(err)
		}
	*/

	var m *golang.Mod
	if _, err := os.Stat("go.mod"); err == nil {
		logrus.Info("found 'go.mod'. Running `go mod edit -json'")
		m, err = golang.LoadGoModFile(".")
		if err != nil {
			logrus.Fatal(err)
		}
	} else if _, err = os.Stat("Cargo.toml"); err == nil {
		// do the thing
	} else {
		logrus.Fatal("no input provided")
	}

	client := gh.New(ctx, os.Getenv("GITHUB_TOKEN"))

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
	return nil
}

type Check struct {
	Import   string
	VcsUrl   *urlpkg.URL
	Archived bool
}

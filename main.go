package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"github.com/vbatts/is-archived/pkg/check"
	"github.com/vbatts/is-archived/pkg/cratesio"
	"github.com/vbatts/is-archived/pkg/gh"
	"github.com/vbatts/is-archived/pkg/golang"
)

func init() {
	flag.Parse()
	logrus.SetOutput(os.Stderr)
}

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

	toCheck := []check.Check{}
	if _, err := os.Stat("go.mod"); err == nil {
		logrus.Info("found 'go.mod'. Running `go mod edit -json'")
		m, err := golang.LoadGoModFile(".")
		if err != nil {
			logrus.Fatal(err)
		}
		c, err := golang.ToCheck(m)
		if err != nil {
			logrus.Fatal(err)
		}
		toCheck = append(toCheck, c...)
	} else if _, err = os.Stat("Cargo.toml"); err == nil {
		// do the thing
		logrus.Info("found 'Cargo.toml'")
		cargo, err := cratesio.LoadCargoFile("Cargo.toml")
		if err != nil {
			logrus.Fatal(err)
		}
		c, err := cratesio.ToCheck(cargo)
		if err != nil {
			logrus.Fatal(err)
		}
		toCheck = append(toCheck, c...)
	} else {
		logrus.Fatal("no input provided")
	}

	client := gh.New(ctx, os.Getenv("GITHUB_TOKEN"))

	logrus.Infof("checking %d github projects ...", len(toCheck))
	for _, check := range toCheck {
		if !strings.HasPrefix(check.VcsUrl.Host, "github.com") {
			continue
		}
		org, repo := gh.OrgRepoFromURL(check.VcsUrl)
		isArchived, err := client.IsRepoArchived(org, repo)
		if err != nil {
			logrus.Error(err)
			continue
		}
		if isArchived {
			check.Archived = true
			fmt.Printf("%q is archived (%s)\n", check.PkgName, check.VcsUrl.String())
		}
	}
	// TODO print a combined report
	return nil
}

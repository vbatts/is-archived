package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"github.com/vbatts/is-archived/pkg/check"
	_ "github.com/vbatts/is-archived/pkg/cratesio"
	"github.com/vbatts/is-archived/pkg/gh"
	_ "github.com/vbatts/is-archived/pkg/golang"
	_ "github.com/vbatts/is-archived/pkg/npm"
	"github.com/vbatts/is-archived/pkg/types"
	"github.com/vbatts/is-archived/version"
)

func main() {
	app := cli.App{
		Name:    "is-archived",
		Usage:   "check your project's dep's upstreams for being archived projects",
		Version: version.Version,
		Before: func(c *cli.Context) error {
			logrus.SetOutput(os.Stderr)
			if os.Getenv("DEBUG") != "" {
				logrus.SetLevel(logrus.DebugLevel)
			}
			if c.Bool("debug") {
				logrus.SetLevel(logrus.DebugLevel)
			}
			return nil
		},
		Action: mainFunc,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "show debug output",
				Value: false,
			},
		},
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
	foundPackagers := false
	for _, ft := range types.PackagerFileTypes() {
		if _, err := os.Stat(ft); err == nil {
			logrus.Infof("found a %q", ft)
			foundPackagers = true
			p, err := types.GetPackager(ft)
			if err != nil {
				logrus.Error(err)
				continue
			}
			checks, err := p.LoadFile(ft)
			if err != nil {
				logrus.Error(err)
				continue
			}
			logrus.Infof("  added %d packages to check", len(checks))
			toCheck = append(toCheck, checks...)
		}
	}
	if !foundPackagers {
		logrus.Fatal("no known packager filetypes found!")
	}

	if _, err := os.Stat("Gemfile"); err == nil {
		logrus.Error("ruby Gemfile not implemented yet")
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

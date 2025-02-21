package main

import (
	"context"
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

func main() {
	app := cli.App{
		Name:  "is-archived",
		Usage: "check your project's dep's upstreams for being archived projects",
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
		c, err := cratesio.ToCheckCargo(cargo)
		if err != nil {
			logrus.Fatal(err)
		}
		toCheck = append(toCheck, c...)
		if _, err = os.Stat("Cargo.lock"); err == nil {
			logrus.Info("also found 'Cargo.lock'")
			cl, err := cratesio.LoadCargoLockFile("Cargo.lock")
			if err != nil {
				logrus.Fatal(err)
			}
			c, err := cratesio.ToCheckCargoLock(cl)
			if err != nil {
				logrus.Fatal(err)
			}
			toCheck = append(toCheck, c...)
		}
	} else if _, err = os.Stat("Gemfile"); err == nil {
		logrus.Error("ruby Gemfile not implemented yet")
	} else if _, err = os.Stat("package.json"); err == nil {
		logrus.Error("npm package.json not implemented yet")
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

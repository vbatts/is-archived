package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"github.com/vbatts/is-archived/pkg/check"
	_ "github.com/vbatts/is-archived/pkg/cratesio"
	_ "github.com/vbatts/is-archived/pkg/gh"
	_ "github.com/vbatts/is-archived/pkg/golang"
	_ "github.com/vbatts/is-archived/pkg/npm"
	_ "github.com/vbatts/is-archived/pkg/pypi"
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
	if _, err := os.Stat("Gemfile"); err == nil {
		logrus.Error("ruby Gemfile not implemented yet")
	}
	if !foundPackagers {
		logrus.Fatal("no known packager filetypes found!")
	}

	// XXX ok let's see, we'll have a list of all repos "toCheck", and they have
	// already been only added to the list if the RepoerDomains() shows the
	// domain as supported.
	// From this we could either extract and group the checks by domain,
	// or just iterate through the checks, needing a way to run the appropriate
	// check for the domain...

	//logrus.Infof("checking %d github projects ...", len(toCheck))
	for _, ck := range toCheck {
		logrus.Debugf("checking %q (%s)", ck.PkgName, ck.VcsUrl.String())
		err := types.RepoerRun(&ck)
		if err != nil {
			logrus.Warnf("for %q: %s", ck.VcsUrl.String(), err)
			continue
		}
		if ck.Archived {
			fmt.Printf("%q is archived (%s)\n", ck.PkgName, ck.VcsUrl.String())
		}
	}
	// TODO print a combined report
	return nil
}

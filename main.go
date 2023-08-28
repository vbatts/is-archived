package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v54/github"
	"github.com/sirupsen/logrus"
)

/*
parse a `go.mod` and query github whether any of the github projects are archived.

Testing using `go mod edit -json > mod.json`.
This could be done as a file, or subshell or on std-in.
*/
func main() {
	buf, err := os.ReadFile("mod.json") // ... do better ...
	if err != nil {
		logrus.Fatal(err)
	}
	m := Mod{}
	err = json.Unmarshal(buf, &m)
	if err != nil {
		logrus.Fatal(err)
	}

	client := github.NewClient(nil)
	for _, req := range m.Require {
		if !strings.HasPrefix(req.Path, "github.com") {
			continue
		}

		spl := strings.Split(req.Path, "/")

		repo, _, err := client.Repositories.Get(context.Background(), spl[1], spl[2])
		if err != nil {
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

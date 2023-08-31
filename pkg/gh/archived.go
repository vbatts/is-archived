package gh

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// New is a helper to create a client with (or without ...) a token
func New(ctx context.Context, token string) *github.Client {
	var client *github.Client
	if token != "" {
		// query Github for each repo
		// needs PAT for rate limiting ...

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}
	return client
}

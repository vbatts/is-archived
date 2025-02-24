package gh

import (
	"context"
	"fmt"
	urlpkg "net/url"
	"strings"

	"github.com/google/go-github/github"
	"github.com/vbatts/is-archived/version"
	"golang.org/x/oauth2"
)

type Handler struct {
	ctx    context.Context
	Client *github.Client
}

// New is a helper to create a client with (or without ...) a token
func New(ctx context.Context, token string) Handler {
	h := Handler{
		ctx: ctx,
	}
	if token != "" {
		// query Github for each repo
		// needs PAT for rate limiting ...

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(h.ctx, ts)
		h.Client = github.NewClient(tc)
	} else {
		h.Client = github.NewClient(nil)
	}
	h.Client.UserAgent = fmt.Sprintf("%s/%s", version.Project, version.Version)
	return h
}

func OrgRepoFromURL(vcsurl *urlpkg.URL) (string, string) {
	spl := strings.Split(vcsurl.Path, "/")
	if len(spl) > 2 {
		return spl[1], spl[2]
	}
	return "", ""
}

func (h *Handler) IsRepoArchived(org, repo string) (bool, error) {
	ghrepo, _, err := h.Client.Repositories.Get(h.ctx, org, repo)
	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return false, fmt.Errorf("rate limited. Try using a Personal Access Token and setting GITHUB_TOKEN env variable: %w", err)
		}
		if _, ok := err.(*github.AbuseRateLimitError); ok {
			return false, fmt.Errorf("rate limited. Try using a Personal Access Token and setting GITHUB_TOKEN env variable: %w", err)
		}
		return false, fmt.Errorf("failed fetching github.com/%s/%s: %w", org, repo, err)
	}
	return ghrepo.GetArchived(), nil
}

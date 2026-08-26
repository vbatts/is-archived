package forgejo

import (
	urlpkg "net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vbatts/is-archived/pkg/check"
	"github.com/vbatts/is-archived/pkg/types"
)

// Name identifies this Repoer's checks in logs.
const Name = "Forgejo/Gitea/Gogs"

func init() {
	types.RegisterRepoer(forgejoRepoer{})
}

// forgejoRepoer is a wildcard Repoer: forgejo, gitea, and gogs are
// self-hosted under arbitrary domains rather than one fixed host, so it is
// tried as a fallback for any host no domain-specific Repoer claimed. See
// types.Repoer.Domain and types.RepoerRun.
type forgejoRepoer struct{}

func (r forgejoRepoer) Domain() string {
	return ""
}

func (r forgejoRepoer) Run(ck *check.Check) error {
	org, repo := orgRepoFromURL(ck.VcsUrl)
	if org == "" || repo == "" {
		return nil
	}

	rr, ok, err := FetchRepo(ck.VcsUrl.Scheme, ck.VcsUrl.Host, org, repo)
	if err != nil {
		return err
	}
	if !ok {
		logrus.Debugf("[%s] %s does not look like a forgejo/gitea/gogs instance", Name, ck.VcsUrl.Host)
		return nil
	}
	ck.Archived = rr.Archived
	return nil
}

func orgRepoFromURL(vcsurl *urlpkg.URL) (string, string) {
	spl := strings.Split(vcsurl.Path, "/")
	if len(spl) > 2 {
		return spl[1], strings.TrimSuffix(spl[2], ".git")
	}
	return "", ""
}

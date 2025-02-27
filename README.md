# is-archived

Check if the repos you're depending on are archived upstream.

Currently checks:
- golang `go.mod`
- rust `Cargo.toml` and `Cargo.lock`
- nodejs `package.json`
- python `requirements.txt`

## Usage

```shell
vbatts@weasel:~/src/github.com/vbatts/is-archived$ is-archived
INFO[0000] found 'go.mod'. Running `go mod edit -json'
INFO[0000] checking 6 github projects ...
vbatts@possibly:~/src/cc/image-rs$ is-archived
INFO[0000] found 'Cargo.toml'                           
INFO[0000] "" does not match "attestation_agent"        
INFO[0001] "" does not match "ocicrypt-rs"              
INFO[0001] "sigstore" does not list a repository        
INFO[0001] checking 28 github projects ...              
vbatts@possibly:~/src/cc/image-rs$ 

```

## Install

```shell
go install github.com/vbatts/is-archived@latest
```

## Github Rate Limit

With even a project like kubernetes, you'll hit the Github rate limit on the first run.

Go create a personal access token (PAT) on your [Github Setting](https://github.com/settings/tokens?type=beta), and export it as a local environment variable.

```shell
export GITHUB_TOKEN=<your_github_pat>
```

## Roadmap Ideas

- [ ] multiple source code services
  - [x] github repo API
  - [ ] gitlab project API (like https://docs.gitlab.com/ee/api/projects.html#get-single-project)
  - [ ] bitbucket project API (like https://developer.atlassian.com/cloud/bitbucket/rest/api-group-repositories/#api-repositories-workspace-repo-slug-get)
  - [ ] gogs/gitea/forgejo API (like https://docs.gitea.com/api/1.23/#tag/repository/operation/repoGet)
- [ ] multiple languages
  - [x] golang `go.mod`
  - [x] javascript / npm `packages.json`
  - [x] python / pypi `requirements.txt`
    - [ ] `pyproject.toml` https://packaging.python.org/en/latest/guides/writing-pyproject-toml/
  - [x] rust `Cargo.toml`
  - [ ] ruby / rubygems `Gemfile`
    - [API](https://guides.rubygems.org/rubygems-org-api/)
    - [Gemfile locks](https://stackoverflow.com/questions/7517524/understanding-the-gemfile-lock-file)
    - [Bundler Lockfile parser](https://github.com/rubygems/rubygems/blob/07e3756fd894e5ded0206bc309dc64ff8ba48f8f/bundler/lib/bundler/lockfile_parser.rb#L4)
  - [ ] who knows?
- [x] golang to pull-through the HTML `<meta name="go-import" ...` redirects
- [ ] detect if stdout is terminal or pipe. If Terminal, then get fancy with [bubbletea](https://github.com/charmbracelet/bubbletea)


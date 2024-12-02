# is-archived

check if the repos you're depending on are archived upstream.

Currently checks a golang `go.mod` file and/or a rust `Cargo.toml`

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

- [x] github repo API
- [ ] gitlab project API (like https://docs.gitlab.com/ee/api/projects.html#get-single-project)
- [ ] bitbucket project API (like https://developer.atlassian.com/cloud/bitbucket/rest/api-group-repositories/#api-repositories-workspace-repo-slug-get)
- [ ] multiple languages
  - [x] golang `go.mod`
  - [ ] javascript `packages.json`
  - [x] rust `Cargo.toml`
  - [ ] rubygems? 
    - [API](https://guides.rubygems.org/rubygems-org-api/)
    - [Gemfile locks](https://stackoverflow.com/questions/7517524/understanding-the-gemfile-lock-file)
    - [Bundler Lockfile parser](https://github.com/rubygems/rubygems/blob/07e3756fd894e5ded0206bc309dc64ff8ba48f8f/bundler/lib/bundler/lockfile_parser.rb#L4)
  - [ ] npm?
    - [NPM registry API](https://www.edoardoscibona.com/exploring-the-npm-registry-api)
  - [ ] who knows?
- [x] golang to pull-through the HTML `<meta name="go-import" ...` redirects
- [ ] detect if stdout is terminal or pipe. If Terminal, then get fancy with [bubbletea](https://github.com/charmbracelet/bubbletea)


# is-archived

check if the repos you're depending on are archived upstream.

Currently checks a golang `go.mod` file.

## Usage

```shell
vbatts@weasel:~/src/github.com/vbatts/is-archived$ is-archived
INFO[0000] found 'go.mod'. Running `go mod edit -json'
INFO[0000] checking 6 github projects ...
vbatts@weasel:~/src/github.com/vbatts/is-archived$ go mod edit -json | is-archived
INFO[0000] reading from stdin ...
INFO[0000] checking 6 github projects ...
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

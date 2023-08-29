# is-archived

check if the repos you're depending on are archived upstream.

## Github Rate Limit

With even a project like kubernetes, you'll hit the Github rate limit on the first run.

Go create a personal access token (PAT) on your [Github Setting](https://github.com/settings/tokens?type=beta), and export it as a local environment variable.

```shell
export GITHUB_TOKEN=<your_github_pat>
```

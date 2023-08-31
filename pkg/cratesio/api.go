package cratesio

/*
curl -sSL "https://crates.io/api/v1/crates?per_page=100&page=1" | \
jq '.crates[] | select(.repository != null) ' | less

curl -sSL "https://crates.io/api/v1/crates/aarch64-esr-decoder" | jq . | less
*/

type Crate struct {
	ID         string `json:"id"`
	Repository string `json:"repository"`
}

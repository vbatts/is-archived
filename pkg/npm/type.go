package npm

// Package is of the `package.json` file
type Package struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Desription           string            `json:"description"`
	Main                 string            `json:"main,omitempty"`
	Scripts              map[string]string `json:"scripts,omitempty"`
	Repository           Repository        `json:"repository,omitempty"`
	License              string            `json:"license,omitempty"`
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies,omitempty"`
	OptionalDependencies map[string]string `json:"optionalDependencies,omitempty"`
	//Author          string            `json:"author,omitempty"` // TODO this can be a string OR a struct ...
}

type Repository struct {
	URL       string
	Type      string
	Directory string
}

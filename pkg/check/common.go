package check

import (
	urlpkg "net/url"
)

// Check is the unit of work to be evaulated, as well as the result of whether
// a package has been archived.
type Check struct {
	Lang     string
	PkgName  string
	VcsUrl   *urlpkg.URL
	Archived bool
}

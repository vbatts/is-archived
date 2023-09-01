package check

import (
	urlpkg "net/url"
)

type Check struct {
	Lang     string
	PkgName  string
	VcsUrl   *urlpkg.URL
	Archived bool
}

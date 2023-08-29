// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vcs

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"strings"
)

// ModuleMode specifies whether to prefer modules when looking up code sources.
type ModuleMode int

const (
	IgnoreMod ModuleMode = iota
	PreferMod
)

// MetaImport represents the parsed <meta name="go-import"
// content="prefix vcs reporoot" /> tags from HTML files.
type MetaImport struct {
	Prefix, VCS, RepoRoot string
}

// charsetReader returns a reader that converts from the given charset to UTF-8.
// Currently it only supports UTF-8 and ASCII. Otherwise, it returns a meaningful
// error which is printed by go get, so the user can find why the package
// wasn't downloaded if the encoding is not supported. Note that, in
// order to reduce potential errors, ASCII is treated as UTF-8 (i.e. characters
// greater than 0x7f are not rejected).
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "utf-8", "ascii":
		return input, nil
	default:
		return nil, fmt.Errorf("can't decode XML document using charset %q", charset)
	}
}

// urlForImportPath returns a partially-populated URL for the given Go import path.
//
// The URL leaves the Scheme field blank so that web.Get will try any scheme
// allowed by the selected security mode.
func urlForImportPath(importPath string) (*urlpkg.URL, error) {
	slash := strings.Index(importPath, "/")
	if slash < 0 {
		slash = len(importPath)
	}
	host, path := importPath[:slash], importPath[slash:]
	if !strings.Contains(host, ".") {
		return nil, errors.New("import path does not begin with hostname")
	}
	if len(path) == 0 {
		path = "/"
	}
	return &urlpkg.URL{Host: host, Path: path, RawQuery: "go-get=1"}, nil
}

func MetaImportForPath(importPath string) (*urlpkg.URL, []MetaImport, error) {
	url, err := urlForImportPath(importPath)
	if err != nil {
		return url, nil, err
	}
	resp, err := http.DefaultClient.Do(&http.Request{
		Method: "GET",
		URL:    url,
	})
	if err != nil {
		return url, nil, fmt.Errorf("fetching %s: %v", importPath, err)
	}
	defer resp.Body.Close()
	imports, err := parseMetaGoImports(resp.Body, PreferMod)
	if err != nil {
		return url, nil, fmt.Errorf("parsing %s: %v", importPath, err)
	}
	if len(imports) == 0 {
		return url, nil, fmt.Errorf("fetching %s: no go-import meta tag found %s", importPath, url)
	}
	return url, imports, nil
}

// parseMetaGoImports returns meta imports from the HTML in r.
// Parsing ends at the end of the <head> section or the beginning of the <body>.
func parseMetaGoImports(r io.Reader, mod ModuleMode) ([]MetaImport, error) {
	d := xml.NewDecoder(r)
	d.CharsetReader = charsetReader
	d.Strict = false
	var imports []MetaImport
	for {
		t, err := d.RawToken()
		if err != nil {
			if err != io.EOF && len(imports) == 0 {
				return nil, err
			}
			break
		}
		if e, ok := t.(xml.StartElement); ok && strings.EqualFold(e.Name.Local, "body") {
			break
		}
		if e, ok := t.(xml.EndElement); ok && strings.EqualFold(e.Name.Local, "head") {
			break
		}
		e, ok := t.(xml.StartElement)
		if !ok || !strings.EqualFold(e.Name.Local, "meta") {
			continue
		}
		if attrValue(e.Attr, "name") != "go-import" {
			continue
		}
		if f := strings.Fields(attrValue(e.Attr, "content")); len(f) == 3 {
			imports = append(imports, MetaImport{
				Prefix:   f[0],
				VCS:      f[1],
				RepoRoot: f[2],
			})
		}
	}

	// Extract mod entries if we are paying attention to them.
	var list []MetaImport
	var have map[string]bool
	if mod == PreferMod {
		have = make(map[string]bool)
		for _, m := range imports {
			if m.VCS == "mod" {
				have[m.Prefix] = true
				list = append(list, m)
			}
		}
	}

	// Append non-mod entries, ignoring those superseded by a mod entry.
	for _, m := range imports {
		if m.VCS != "mod" && !have[m.Prefix] {
			list = append(list, m)
		}
	}
	return list, nil
}

// attrValue returns the attribute value for the case-insensitive key
// `name', or the empty string if nothing is found.
func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}

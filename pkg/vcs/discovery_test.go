// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vcs

import (
	"reflect"
	"strings"
	"testing"
)

var parseMetaGoImportsTests = []struct {
	in  string
	mod ModuleMode
	out []MetaImport
}{
	{
		`<meta name="go-import" content="foo/bar git https://github.com/rsc/foo/bar">`,
		IgnoreMod,
		[]MetaImport{{"foo/bar", "git", "https://github.com/rsc/foo/bar"}},
	},
	{
		`<meta name="go-import" content="foo/bar git https://github.com/rsc/foo/bar">
		<meta name="go-import" content="baz/quux git http://github.com/rsc/baz/quux">`,
		IgnoreMod,
		[]MetaImport{
			{"foo/bar", "git", "https://github.com/rsc/foo/bar"},
			{"baz/quux", "git", "http://github.com/rsc/baz/quux"},
		},
	},
	{
		`<meta name="go-import" content="foo/bar git https://github.com/rsc/foo/bar">
		<meta name="go-import" content="foo/bar mod http://github.com/rsc/baz/quux">`,
		IgnoreMod,
		[]MetaImport{
			{"foo/bar", "git", "https://github.com/rsc/foo/bar"},
		},
	},
	{
		`<meta name="go-import" content="foo/bar mod http://github.com/rsc/baz/quux">
		<meta name="go-import" content="foo/bar git https://github.com/rsc/foo/bar">`,
		IgnoreMod,
		[]MetaImport{
			{"foo/bar", "git", "https://github.com/rsc/foo/bar"},
		},
	},
	{
		`<meta name="go-import" content="foo/bar mod http://github.com/rsc/baz/quux">
		<meta name="go-import" content="foo/bar git https://github.com/rsc/foo/bar">`,
		PreferMod,
		[]MetaImport{
			{"foo/bar", "mod", "http://github.com/rsc/baz/quux"},
		},
	},
	{
		`<head>
		<meta name="go-import" content="foo/bar git https://github.com/rsc/foo/bar">
		</head>`,
		IgnoreMod,
		[]MetaImport{{"foo/bar", "git", "https://github.com/rsc/foo/bar"}},
	},
	{
		`<meta name="go-import" content="foo/bar git https://github.com/rsc/foo/bar">
		<body>`,
		IgnoreMod,
		[]MetaImport{{"foo/bar", "git", "https://github.com/rsc/foo/bar"}},
	},
	{
		`<!doctype html><meta name="go-import" content="foo/bar git https://github.com/rsc/foo/bar">`,
		IgnoreMod,
		[]MetaImport{{"foo/bar", "git", "https://github.com/rsc/foo/bar"}},
	},
	{
		// XML doesn't like <div style=position:relative>.
		`<!doctype html><title>Page Not Found</title><meta name=go-import content="chitin.io/chitin git https://github.com/chitin-io/chitin"><div style=position:relative>DRAFT</div>`,
		IgnoreMod,
		[]MetaImport{{"chitin.io/chitin", "git", "https://github.com/chitin-io/chitin"}},
	},
	{
		`<meta name="go-import" content="myitcv.io git https://github.com/myitcv/x">
	        <meta name="go-import" content="myitcv.io/blah2 mod https://raw.githubusercontent.com/myitcv/pubx/master">
	        `,
		IgnoreMod,
		[]MetaImport{{"myitcv.io", "git", "https://github.com/myitcv/x"}},
	},
	{
		`<meta name="go-import" content="myitcv.io git https://github.com/myitcv/x">
	        <meta name="go-import" content="myitcv.io/blah2 mod https://raw.githubusercontent.com/myitcv/pubx/master">
	        `,
		PreferMod,
		[]MetaImport{
			{"myitcv.io/blah2", "mod", "https://raw.githubusercontent.com/myitcv/pubx/master"},
			{"myitcv.io", "git", "https://github.com/myitcv/x"},
		},
	},
}

func TestParseMetaGoImports(t *testing.T) {
	for i, tt := range parseMetaGoImportsTests {
		out, err := parseMetaGoImports(strings.NewReader(tt.in), tt.mod)
		if err != nil {
			t.Errorf("test#%d: %v", i, err)
			continue
		}
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("test#%d:\n\thave %q\n\twant %q", i, out, tt.out)
		}
	}
}

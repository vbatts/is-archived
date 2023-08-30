package vcs

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var goOpenCensusIoHtml = `
<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="go.opencensus.io git https://github.com/census-instrumentation/opencensus-go">
<meta name="go-source" content="go.opencensus.io https://github.com/census-instrumentation/opencensus-go https://github.com/census-instrumentation/opencensus-go/tree/master{/dir} https://github.com/census-instrumentation/opencensus-go/blob/master{/dir}/{file}#L{line}">
<meta http-equiv="refresh" content="0; url=https://pkg.go.dev/go.opencensus.io/">
</head>
<body>
Nothing to see here; <a href="https://pkg.go.dev/go.opencensus.io/">see the package on pkg.go.dev</a>.
</body>
</html>
`

type serve struct {
}

func (s serve) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, goOpenCensusIoHtml)
}

func TestMetaImportForPath(t *testing.T) {
	server := httptest.NewServer(serve{})
	defer server.Close()
	u1, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	u2, mi, err := MetaImportForPath(u1.Host)
	if err != nil {
		t.Fatal(err)
	}

	if u1.Host != u2.Host {
		t.Errorf("expected %q to match %q", u1, u2)
	}

	u3, err := url.Parse(mi[0].RepoRoot)
	if err != nil {
		t.Fatal(err)
	}

	if u3.Host != "github.com" {
		t.Fatalf("expected %q to equal 'github.com'", u3.Host)
	}
	if u3.Path != "/census-instrumentation/opencensus-go" {
		t.Fatalf("expected %q to equal '/census-instrumentation/opencensus-go'", u3.Path)
	}
}

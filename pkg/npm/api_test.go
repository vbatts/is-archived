package npm

import (
	"net/url"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestEndpoint(t *testing.T) {
	isEndpointAvailable()

	p, err := FetchSingle("help")
	if err != nil {
		t.Error(err)
	}
	u, err := url.Parse(p.Repository.URL)
	if err != nil {
		t.Error(err)
	}
	logrus.Infof("%s://%s%s", u.Scheme, u.Host, u.Path)

}

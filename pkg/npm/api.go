package npm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/vbatts/is-archived/version"
)

// https://www.edoardoscibona.com/exploring-the-npm-registry-api
// https://github.com/npm/registry/blob/main/docs/REGISTRY-API.md

// this endpoint is deprecated
// const apiEndpoint = "https://registry.npmjs.org/"
const apiEndpoint = "https://replicate.npmjs.com"

func isEndpointAvailable() bool {
	resp, err := http.Head(apiEndpoint)
	if err != nil {
		logrus.Errorf("isEndpointAvailable: %s", err)
		return false
	}

	logrus.Infof("%#v", resp)
	if resp.StatusCode == 200 {
		return true
	}
	return false
}

func FetchSingle(pkg string) (*Package, error) {
	u := fmt.Sprintf("%s/%s", apiEndpoint, pkg)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching %q: %w", u, err)
	}
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", version.Project, version.Version))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//logrus.Infof("%#v", string(buf))

	p := Package{}
	err = json.Unmarshal(buf, &p)
	if err != nil {
		return nil, err
	}

	//logrus.Info(p.Repository.URL)
	return &p, nil
}

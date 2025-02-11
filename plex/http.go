package plex

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"mime"
	"net/http"
)

// httpRequest sends a HTTP request according to provided method and url,
// decoding the response as JSON or XML
func httpRequest[V any](client *http.Client, method string, url string, headers map[string]string) (*V, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d for url %s", resp.StatusCode, req.URL.String())
	}

	t, _, err := mime.ParseMediaType(resp.Header.Get("content-type"))

	var parsed V
	switch t {
	case "application/json":
		err = json.NewDecoder(resp.Body).Decode(&parsed)
	case "application/xml":
		err = xml.NewDecoder(resp.Body).Decode(&parsed)
	default:
		err = fmt.Errorf("unexpected content-type: %s", t)
	}
	return &parsed, err
}

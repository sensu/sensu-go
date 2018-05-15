package slack

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
)

var httpClient *http.Client

func init() {
	httpClient = &http.Client{}
}

func (sl *Slack) request(req *http.Request) ([]byte, error) {
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

func (sl *Slack) GetRequest(endpoint string, uv *url.Values) ([]byte, error) {
	ul := apiBaseUrl + endpoint
	req, err := http.NewRequest("GET", ul, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = (*uv).Encode()
	return sl.request(req)
}

func (sl *Slack) PostRequest(endpoint string, uv *url.Values, body *bytes.Buffer) ([]byte, error) {
	ul := apiBaseUrl + endpoint
	req, err := http.NewRequest("POST", ul, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = (*uv).Encode()
	return sl.request(req)
}

func (sl *Slack) DoRequest(req *http.Request) ([]byte, error) {
	return sl.request(req)
}

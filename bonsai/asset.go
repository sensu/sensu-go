package bonsai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// FetchAsset fetches an asset (list of versions)
func (c *RestClient) FetchAsset(namespace, name string) (*Asset, error) {
	req, err := c.newGetRequest(namespace, name)
	if err != nil {
		return nil, err
	}

	logger.WithField("request", req.URL.String()).Debug("sending request to bonsai")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if code := resp.StatusCode; code >= 400 {
		return nil, fmt.Errorf("bonsai api returned status code: %d", code)
	}

	var asset Asset
	if err := json.NewDecoder(resp.Body).Decode(&asset); err != nil {
		return nil, err
	}

	return &asset, nil
}

// FetchAssetVersion fetches an asset definition for a the specified asset version
func (c *RestClient) FetchAssetVersion(namespace, name, version string) (string, error) {
	req, err := c.newGetRequest(namespace, name, version, "release_asset_builds")
	if err != nil {
		return "", err
	}

	logger.WithField("request", req.URL.String()).Debug("sending request to bonsai")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if code := resp.StatusCode; code >= 400 {
		err = fmt.Errorf("bonsai api returned status code: %d", code)
		return "", err
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return "", err
	}

	return buf.String(), nil
}

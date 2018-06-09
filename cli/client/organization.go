package client

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sensu/sensu-go/types"
)

// CreateOrganization creates new org on configured Sensu instance
func (client *RestClient) CreateOrganization(org *types.Organization) error {
	bytes, err := json.Marshal(org)
	if err != nil {
		return err
	}

	res, err := client.R().
		SetBody(bytes).
		Post("/rbac/organizations")

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// UpdateOrganization updates given organization on a configured Sensu instance
func (client *RestClient) UpdateOrganization(org *types.Organization) error {
	bytes, err := json.Marshal(org)
	if err != nil {
		return err
	}

	res, err := client.R().SetBody(bytes).Put("/rbac/organizations/" + url.PathEscape(org.Name))
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// DeleteOrganization deletes an organization on configured Sensu instance
func (client *RestClient) DeleteOrganization(org string) error {
	res, err := client.R().Delete("/rbac/organizations/" + url.PathEscape(org))

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// ListOrganizations fetches all organizations from configured Sensu instance
func (client *RestClient) ListOrganizations() ([]types.Organization, error) {
	var orgs []types.Organization

	res, err := client.R().Get("/rbac/organizations")
	if err != nil {
		return orgs, err
	}

	if res.StatusCode() >= 400 {
		return orgs, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &orgs)
	return orgs, err
}

// FetchOrganization fetches an organization by name
func (client *RestClient) FetchOrganization(orgName string) (*types.Organization, error) {
	var org *types.Organization

	res, err := client.R().Get("/rbac/organizations/" + url.PathEscape(orgName))
	if err != nil {
		return org, err
	}

	if res.StatusCode() >= 400 {
		return org, fmt.Errorf("error getting organization: %v", res.String())
	}

	err = json.Unmarshal(res.Body(), &org)
	return org, err
}

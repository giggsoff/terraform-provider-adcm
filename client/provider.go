package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetProviders - Returns list of providers
func (c *Client) GetProviders() ([]Provider, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/provider", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}

	var ids []Identifier
	err = json.Unmarshal(body, &ids)
	if err != nil {
		return nil, err
	}

	var providers []Provider
	for _, id := range ids {
		req, err = http.NewRequest("GET", fmt.Sprintf("%s/api/v1/provider/%d", c.HostURL, id.ID), nil)
		if err != nil {
			return nil, err
		}
		body, err := c.doRequest(req, nil)
		if err != nil {
			return nil, err
		}
		var provider Provider
		err = json.Unmarshal(body, &provider)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}

	return providers, nil
}

// GetProvider - Returns provider
func (c *Client) GetProvider(searchOpts ProviderSearch) (*Provider, error) {
	bundles, err := c.GetProviders()
	if err != nil {
		return nil, err
	}
	var res []Provider
	for _, b := range bundles {
		if searchOpts.Name != "" && searchOpts.Name != b.Name {
			continue
		}
		if searchOpts.Description != "" && searchOpts.Description != b.Description {
			continue
		}
		if searchOpts.BundleID != 0 && searchOpts.BundleID != b.BundleID {
			continue
		}
		if searchOpts.State != "" && searchOpts.State != b.State {
			continue
		}
		if searchOpts.ID != 0 && searchOpts.ID != b.ID {
			continue
		}
		res = append(res, b)
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("your query returned no results. Please change your search criteria and try again")
	}
	if len(res) > 1 {
		return nil, fmt.Errorf("your query returned more than one result. Please try a more specific search criteria")
	}
	return &res[0], nil
}

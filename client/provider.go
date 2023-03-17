package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/imdario/mergo"
	"net/http"
)

func (c *Client) getProviderPrototypeID(bundleID int64) (int64, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/stack/provider/?bundle_id=%d", c.HostURL, bundleID), nil)
	if err != nil {
		return 0, err
	}
	body, err := c.doRequest(req, nil)
	if err != nil {
		return 0, err
	}
	var clusterPrototypeIDS []Identifier
	err = unwrapResults(body, &clusterPrototypeIDS)
	if err != nil {
		return 0, err
	}
	if len(clusterPrototypeIDS) < 1 {
		return 0, fmt.Errorf("no cluster prototypes found")
	}
	return clusterPrototypeIDS[0].ID, nil
}

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

// CreateProvider - create provider
func (c *Client) CreateProvider(provider Provider) (*Provider, error) {
	providerPrototypeID, err := c.getProviderPrototypeID(provider.BundleID)
	if err != nil {
		return nil, err
	}

	values := map[string]interface{}{"name": provider.Name, "description": provider.Description, "prototype_id": providerPrototypeID}
	jsonValue, _ := json.Marshal(values)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/provider/", c.HostURL), bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	body, err := c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}
	var clusterID Identifier
	err = json.Unmarshal(body, &clusterID)
	if err != nil {
		return nil, err
	}
	if len(provider.ProviderConfig.Config) > 0 {
		createdProvider, err := c.GetProvider(ProviderSearch{Identifier: clusterID})
		if err != nil {
			return nil, err
		}
		err = mergo.Merge(&createdProvider.ProviderConfig.Config, provider.ProviderConfig.Config, mergo.WithOverride)
		if err != nil {
			return nil, err
		}
		cfgResponse := ProviderConfigResponse{createdProvider.ProviderConfig.Config}
		data, err := json.Marshal(cfgResponse)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/provider/%d/config/history/", c.HostURL, clusterID.ID), bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
		_, err = c.doRequest(req, nil)
		if err != nil {
			return nil, err
		}
	}

	return c.GetProvider(ProviderSearch{Identifier: clusterID})
}

// DeleteProvider - create host
func (c *Client) DeleteProvider(provider ProviderSearch) error {
	h, err := c.GetProvider(provider)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/provider/%d/", c.HostURL, h.ID), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, nil)
	if err != nil {
		return err
	}
	return nil
}

package client

import (
	"fmt"
	"net/http"
)

// GetBundles - Returns list of bundles
func (c *Client) GetBundles() ([]Bundle, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/stack/bundle", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}

	var bundles []Bundle
	err = unwrapResults(body, &bundles)
	if err != nil {
		return nil, err
	}

	return bundles, nil
}

// GetBundle - Returns bundle
func (c *Client) GetBundle(searchOpts BundleSearch) (*Bundle, error) {
	bundles, err := c.GetBundles()
	if err != nil {
		return nil, err
	}
	var res []Bundle
	for _, b := range bundles {
		if searchOpts.Name != "" && searchOpts.Name != b.Name {
			continue
		}
		if searchOpts.Description != "" && searchOpts.Description != b.Description {
			continue
		}
		if searchOpts.Version != "" && searchOpts.Version != b.Version {
			continue
		}
		if searchOpts.License != "" && searchOpts.License != b.License {
			continue
		}
		if searchOpts.DisplayName != "" && searchOpts.DisplayName != b.DisplayName {
			continue
		}
		if searchOpts.Edition != "" && searchOpts.Edition != b.Edition {
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

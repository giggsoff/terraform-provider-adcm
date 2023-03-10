package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/imdario/mergo"
)

// CreateHost - create host
func (c *Client) CreateHost(host Host) (*Host, error) {
	values := map[string]string{"fqdn": host.FQDN, "description": host.Description}

	jsonValue, _ := json.Marshal(values)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/provider/%d/host/", c.HostURL, host.ProviderID), bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	body, err := c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}
	var id Identifier
	err = json.Unmarshal(body, &id)
	if err != nil {
		return nil, err
	}
	h, err := c.GetHost(HostSearch{Identifier: id})
	if err != nil {
		return nil, err
	}
	if len(host.Config) > 0 {
		err := mergo.Merge(&h.Config, host.Config, mergo.WithOverride)
		if err != nil {
			return nil, err
		}
		cfgResponse := HostConfigResponse{h.Config}
		data, err := json.Marshal(cfgResponse)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/host/%d/config/history/", c.HostURL, id.ID), bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
		_, err = c.doRequest(req, nil)
		if err != nil {
			return nil, err
		}
	}

	return c.GetHost(HostSearch{Identifier: id})
}

// GetHosts - list host
func (c *Client) GetHosts() ([]Host, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/host/", c.HostURL), nil)
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

	var hosts []Host
	for _, id := range ids {
		req, err = http.NewRequest("GET", fmt.Sprintf("%s/api/v1/host/%d", c.HostURL, id.ID), nil)
		if err != nil {
			return nil, err
		}
		body, err := c.doRequest(req, nil)
		if err != nil {
			return nil, err
		}
		var hostResponse HostResponse
		err = json.Unmarshal(body, &hostResponse)
		if err != nil {
			return nil, err
		}
		var host Host
		host.HostResponse = hostResponse
		req, err = http.NewRequest("GET", fmt.Sprintf("%s/api/v1/host/%d/config/current/", c.HostURL, id.ID), nil)
		if err != nil {
			return nil, err
		}
		body, err = c.doRequest(req, nil)
		if err != nil {
			return nil, err
		}
		var hostConfigResponse HostConfigResponse
		err = json.Unmarshal(body, &hostConfigResponse)
		if err != nil {
			return nil, fmt.Errorf("%s : %s", body, err)
		}
		host.HostConfigResponse = hostConfigResponse
		hosts = append(hosts, host)
	}

	return hosts, nil
}

// GetHost - get host
func (c *Client) GetHost(searchOpts HostSearch) (*Host, error) {
	hosts, err := c.GetHosts()
	if err != nil {
		return nil, err
	}
	var res []Host
	for _, h := range hosts {
		if searchOpts.FQDN != "" && searchOpts.FQDN != h.FQDN {
			continue
		}
		if searchOpts.Description != "" && searchOpts.Description != h.Description {
			continue
		}
		if searchOpts.ProviderID != 0 && searchOpts.ProviderID != h.ProviderID {
			continue
		}
		if searchOpts.ClusterID != 0 && searchOpts.ClusterID != h.ClusterID {
			continue
		}
		if searchOpts.ID != 0 && searchOpts.ID != h.ID {
			continue
		}
		res = append(res, h)
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("your query returned no results. Please change your search criteria and try again")
	}
	if len(res) > 1 {
		return nil, fmt.Errorf("your query returned more than one result. Please try a more specific search criteria")
	}
	return &res[0], nil
}

// DeleteHost - create host
func (c *Client) DeleteHost(host HostSearch) error {
	h, err := c.GetHost(host)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/host/%d/", c.HostURL, h.ID), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, nil)
	if err != nil {
		return err
	}
	return nil
}

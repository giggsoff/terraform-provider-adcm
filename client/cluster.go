package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/imdario/mergo"
)

// CreateCluster - create cluster
func (c *Client) CreateCluster(cluster Cluster) (*Cluster, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/stack/cluster/?bundle_id=%d", c.HostURL, cluster.BundleID), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}
	var clusterPrototypeIDS []Identifier
	err = unwrapResults(body, &clusterPrototypeIDS)
	if err != nil {
		return nil, err
	}
	if len(clusterPrototypeIDS) < 1 {
		return nil, fmt.Errorf("no cluster prototypes found")
	}

	req, err = http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/stack/prototype/%d/accept_license/", c.HostURL, clusterPrototypeIDS[0].ID), nil)
	if err != nil {
		return nil, err
	}
	_, err = c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}

	values := map[string]interface{}{"name": cluster.Name, "description": cluster.Description, "prototype_id": clusterPrototypeIDS[0].ID}
	jsonValue, _ := json.Marshal(values)
	req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/cluster/", c.HostURL), bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	body, err = c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}
	var id Identifier
	err = json.Unmarshal(body, &id)
	if err != nil {
		return nil, err
	}
	h, err := c.GetCluster(ClusterSearch{Identifier: id})
	if err != nil {
		return nil, err
	}
	if len(cluster.Config) > 0 {
		err := mergo.Merge(&h.Config, cluster.Config, mergo.WithOverride)
		if err != nil {
			return nil, err
		}
		cfgResponse := HostConfigResponse{h.Config}
		data, err := json.Marshal(cfgResponse)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/config/%d/config/history/", c.HostURL, id.ID), bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
		_, err = c.doRequest(req, nil)
		if err != nil {
			return nil, err
		}
	}

	return c.GetCluster(ClusterSearch{Identifier: id})
}

// GetClusters - list clusters
func (c *Client) GetClusters() ([]Cluster, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/cluster/", c.HostURL), nil)
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

	var hosts []Cluster
	for _, id := range ids {
		req, err = http.NewRequest("GET", fmt.Sprintf("%s/api/v1/cluster/%d", c.HostURL, id.ID), nil)
		if err != nil {
			return nil, err
		}
		body, err := c.doRequest(req, nil)
		if err != nil {
			return nil, err
		}
		var clusterResponse ClusterResponse
		err = json.Unmarshal(body, &clusterResponse)
		if err != nil {
			return nil, err
		}
		var host Cluster
		host.ClusterResponse = clusterResponse
		req, err = http.NewRequest("GET", fmt.Sprintf("%s/api/v1/cluster/%d/config/current/", c.HostURL, id.ID), nil)
		if err != nil {
			return nil, err
		}
		body, err = c.doRequest(req, nil)
		if err != nil {
			return nil, err
		}
		var clusterConfigResponse ClusterConfigResponse
		err = json.Unmarshal(body, &clusterConfigResponse)
		if err != nil {
			return nil, fmt.Errorf("%s : %s", body, err)
		}
		host.ClusterConfigResponse = clusterConfigResponse
		hosts = append(hosts, host)
	}

	return hosts, nil
}

// GetCluster - get cluster
func (c *Client) GetCluster(searchOpts ClusterSearch) (*Cluster, error) {
	hosts, err := c.GetClusters()
	if err != nil {
		return nil, err
	}
	var res []Cluster
	for _, h := range hosts {
		if searchOpts.Name != "" && searchOpts.Name != h.Name {
			continue
		}
		if searchOpts.Description != "" && searchOpts.Description != h.Description {
			continue
		}
		if searchOpts.BundleID != 0 && searchOpts.BundleID != h.BundleID {
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

// DeleteCluster - delete cluster
func (c *Client) DeleteCluster(cluster ClusterSearch) error {
	h, err := c.GetCluster(cluster)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/cluster/%d/", c.HostURL, h.ID), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, nil)
	if err != nil {
		return err
	}
	return nil
}

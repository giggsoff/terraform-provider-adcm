package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/imdario/mergo"
)

func (c *Client) getClusterPrototypeID(bundleID int64) (int64, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/stack/cluster/?bundle_id=%d", c.HostURL, bundleID), nil)
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

func (c *Client) getServicePrototypeID(bundleID int64, serviceName string) (int64, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/stack/service/?bundle_id=%d&name=%s", c.HostURL, bundleID, serviceName), nil)
	if err != nil {
		return 0, err
	}
	body, err := c.doRequest(req, nil)
	if err != nil {
		return 0, err
	}
	var servicePrototypeIDS []Identifier
	err = unwrapResults(body, &servicePrototypeIDS)
	if err != nil {
		return 0, err
	}
	if len(servicePrototypeIDS) < 1 {
		return 0, fmt.Errorf("no service prototypes found")
	}
	return servicePrototypeIDS[0].ID, nil
}

func (c *Client) getServiceComponentID(clusterID, serviceID int64, componentName string) (int64, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/api/v1/cluster/%d/service/%d/component/",
			c.HostURL, clusterID, serviceID), nil)
	if err != nil {
		return 0, err
	}
	body, err := c.doRequest(req, nil)
	if err != nil {
		return 0, err
	}
	var components []Component
	err = json.Unmarshal(body, &components)
	if err != nil {
		return 0, err
	}
	for _, el := range components {
		if el.Name == componentName {
			return el.ID, nil
		}
	}
	return 0, fmt.Errorf("no service component id found")
}

func (c *Client) getServiceConfig(clusterID, serviceID int64) (*ServiceConfigResponse, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/api/v1/cluster/%d/service/%d/config/current/",
			c.HostURL, clusterID, serviceID), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}
	var config ServiceConfigResponse
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// CreateCluster - create cluster
func (c *Client) CreateCluster(cluster Cluster) (*Cluster, error) {
	clusterPrototypeID, err := c.getClusterPrototypeID(cluster.BundleID)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/stack/prototype/%d/accept_license/", c.HostURL, clusterPrototypeID), nil)
	if err != nil {
		return nil, err
	}
	_, err = c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}

	values := map[string]interface{}{"name": cluster.Name, "description": cluster.Description, "prototype_id": clusterPrototypeID}
	jsonValue, _ := json.Marshal(values)
	req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/cluster/", c.HostURL), bytes.NewBuffer(jsonValue))
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
	if len(cluster.ClusterConfig.Config) > 0 {
		createdCluster, err := c.GetCluster(ClusterSearch{Identifier: clusterID})
		if err != nil {
			return nil, err
		}
		err = mergo.Merge(&createdCluster.ClusterConfig.Config, cluster.ClusterConfig.Config, mergo.WithOverride)
		if err != nil {
			return nil, err
		}
		cfgResponse := HostConfigResponse{createdCluster.ClusterConfig.Config}
		data, err := json.Marshal(cfgResponse)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/cluster/%d/config/history/", c.HostURL, clusterID.ID), bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
		_, err = c.doRequest(req, nil)
		if err != nil {
			return nil, err
		}
	}
	if len(cluster.HCMap) > 0 {
		var hc []map[string]int64
		addedServices := make(map[string]int64)
		for hostFQDN, serviceList := range cluster.HCMap {
			host, err := c.GetHost(HostSearch{FQDN: hostFQDN})
			if err != nil {
				return nil, err
			}
			jsonValue, _ = json.Marshal(map[string]interface{}{"host_id": host.ID, "description": ""})
			req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/cluster/%d/host/", c.HostURL, clusterID.ID), bytes.NewBuffer(jsonValue))
			if err != nil {
				return nil, err
			}
			req.Header.Add("Content-Type", "application/json;charset=utf-8")
			body, err = c.doRequest(req, nil)
			if err != nil {
				return nil, err
			}
			for _, servicesMapConfig := range serviceList {
				for serviceName, serviceComponents := range servicesMapConfig {
					servicePrototypeID, err := c.getServicePrototypeID(cluster.BundleID, serviceName)
					if err != nil {
						return nil, err
					}
					if _, added := addedServices[serviceName]; !added {
						values := map[string]interface{}{"cluster_id": clusterID.ID, "prototype_id": servicePrototypeID}
						jsonValue, _ := json.Marshal(values)
						req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/cluster/%d/service/", c.HostURL, clusterID.ID), bytes.NewBuffer(jsonValue))
						if err != nil {
							return nil, err
						}
						req.Header.Add("Content-Type", "application/json;charset=utf-8")
						body, err = c.doRequest(req, nil)
						if err != nil {
							return nil, err
						}
						time.Sleep(time.Second)
						var serviceID Identifier
						err = json.Unmarshal(body, &serviceID)
						if err != nil {
							return nil, err
						}
						if val, ok := cluster.ServicesConfig.Config[serviceName]; ok {
							cfgReceived := val.(map[string]interface{})
							cfg, err := c.getServiceConfig(clusterID.ID, serviceID.ID)
							if err != nil {
								return nil, err
							}
							err = mergo.Merge(&cfg.Config, cfgReceived, mergo.WithOverride)
							if err != nil {
								return nil, err
							}
							data, err := json.Marshal(cfg)
							if err != nil {
								return nil, err
							}
							req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/cluster/%d/service/%d/config/history/", c.HostURL, clusterID.ID, serviceID.ID), bytes.NewBuffer(data))
							if err != nil {
								return nil, err
							}
							req.Header.Add("Content-Type", "application/json;charset=utf-8")
							_, err = c.doRequest(req, nil)
							if err != nil {
								return nil, err
							}
						}
						addedServices[serviceName] = serviceID.ID
					}
					for _, componentName := range serviceComponents {
						componentID, err := c.getServiceComponentID(clusterID.ID, addedServices[serviceName], componentName)
						if err != nil {
							return nil, err
						}
						hc = append(hc, map[string]int64{"component_id": componentID, "host_id": host.ID, "service_id": addedServices[serviceName]})
					}
				}
			}
		}
		values := map[string]interface{}{"hc": hc}
		jsonValue, _ := json.Marshal(values)
		req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/cluster/%d/hostcomponent/", c.HostURL, clusterID.ID), bytes.NewBuffer(jsonValue))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
		body, err = c.doRequest(req, nil)
		if err != nil {
			return nil, err
		}
	}

	return c.GetCluster(ClusterSearch{Identifier: clusterID})
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
		host.ClusterConfig = clusterConfigResponse
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

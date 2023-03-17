package client

type Identifier struct {
	ID int64 `json:"id"`
}

type Bundle struct {
	BundleSearch
}

type BundleSearch struct {
	Identifier
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Edition     string `json:"edition"`
	License     string `json:"license"`
	Version     string `json:"version"`
}

type Provider struct {
	ProviderSearch
	ProviderConfig ProviderConfigResponse
}

type ProviderSearch struct {
	Identifier
	Name        string `json:"name"`
	BundleID    int64  `json:"bundle_id"`
	Description string `json:"description"`
	State       string `json:"state"`
}
type ProviderConfigResponse struct {
	Config map[string]interface{} `json:"config"`
}

type Host struct {
	HostResponse
	HostConfigResponse
}

type HostResponse struct {
	HostSearch
}

type HostConfigResponse struct {
	Config map[string]interface{} `json:"config"`
}

type HostSearch struct {
	Identifier
	FQDN        string `json:"fqdn"`
	Description string `json:"description"`
	ProviderID  int64  `json:"provider_id"`
	ClusterID   int64  `json:"cluster_id"`
}

type Cluster struct {
	ClusterResponse
	ServicesConfig ServiceConfigResponse
	ClusterConfig  ClusterConfigResponse
	HCMap          map[string][]map[string][]string `json:"hc_map"`
	State          string                           `json:"state"`
}

type ClusterResponse struct {
	ClusterSearch
}

type ClusterConfigResponse struct {
	Config map[string]interface{} `json:"config"`
}

type ServiceConfigResponse struct {
	Config map[string]interface{} `json:"config"`
	Attr   map[string]interface{} `json:"attr"`
}

type ClusterSearch struct {
	Identifier
	Name        string `json:"name"`
	Description string `json:"description"`
	BundleID    int64  `json:"bundle_id"`
}

type Component struct {
	Identifier
	Name string `json:"name"`
}

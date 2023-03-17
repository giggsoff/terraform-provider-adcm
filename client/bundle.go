package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
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

func (c *Client) UploadBundle(url string) (*Bundle, error) {
	bundleFileName := path.Base(url)
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	r, w := io.Pipe()
	m := multipart.NewWriter(w)
	go func() {
		defer w.Close()
		defer m.Close()
		defer response.Body.Close()
		part, err := m.CreateFormFile("file", bundleFileName)
		if err != nil {
			return
		}
		if _, err = io.Copy(part, response.Body); err != nil {
			return
		}
	}()
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/stack/upload/", c.HostURL), r)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", m.FormDataContentType())

	_, err = c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(map[string]interface{}{"bundle_file": bundleFileName})
	req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/stack/load/", c.HostURL), bytes.NewBuffer(data))
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
	return c.GetBundle(BundleSearch{Identifier: id})
}

// DeleteBundle - Delete bundle
func (c *Client) DeleteBundle(searchOpts BundleSearch) error {
	bundle, err := c.GetBundle(searchOpts)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/stack/bundle/%d/", c.HostURL, bundle.ID), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, nil)
	if err != nil {
		return err
	}
	return nil
}

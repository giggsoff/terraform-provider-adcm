package client

import "testing"

func TestUnwrap(t *testing.T) {
	receivedJson := `
{
	"count": 3,
	"next": null,
	"previous": null,
	"results": [{
		"id": 2,
		"name": "Cloud bundle",
		"version": "14.3",
		"edition": "community",
		"license": "absent",
		"hash": "31047caa9f6e31b651241b89fe6912b612721842",
		"description": "Cloud bundle",
		"date": "2023-03-09T14:18:02.481374Z",
		"license_url": "http://127.0.0.1:8000/api/v1/stack/bundle/2/license/?format=json",
		"update": "http://127.0.0.1:8000/api/v1/stack/bundle/2/update/?format=json",
		"url": "http://127.0.0.1:8000/api/v1/stack/bundle/2/?format=json",
		"adcm_min_version": "2022.07.12.19",
		"display_name": "Cloud bundle"
	}, {
		"id": 3,
		"name": "ADPG",
		"version": "14.4",
		"edition": "enterprise",
		"license": "accepted",
		"hash": "f546ee2ec7b5a476d2c43a75034e4221824d6e04",
		"description": "ADPG",
		"date": "2023-03-09T14:18:58.460360Z",
		"license_url": "http://127.0.0.1:8000/api/v1/stack/bundle/3/license/?format=json",
		"update": "http://127.0.0.1:8000/api/v1/stack/bundle/3/update/?format=json",
		"url": "http://127.0.0.1:8000/api/v1/stack/bundle/3/?format=json",
		"adcm_min_version": "2022.04.18.13",
		"display_name": "ADPG"
	}]
}`
	var bundles []Bundle
	err := unwrapResults([]byte(receivedJson), &bundles)
	if err != nil {
		t.Error(err)
	}
	if len(bundles) != 2 {
		t.Error("Unexpected results count")
	}
}

package adcm

import (
	"fmt"
	"strconv"
	"testing"

	adcmClient "github.com/giggsoff/terraform-provider-adcm/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBundle_basic(t *testing.T) {
	var resName = "adcm_bundle.provider"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderServers,
		PreCheck:                 func() { testAccPreCheck(t) },
		//CheckDestroy:             func(s *terraform.State) error { return destroyADCM(ADCM_ID) },
		Steps: []resource.TestStep{
			{
				Config: getBundleConfig(),
				Check:  testAccCheckBundleExists(resName),
			},
		},
	})
}

func getBundleConfig() string {
	return fmt.Sprintf(`
resource "adcm_bundle" "provider" {
	url = "%s"
}`, bundleStoreAddr+"/provider_v1.0_community.tgz")
}

func testAccCheckBundleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource %s not found", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		client, err := adcmClient.NewClient(&ADCM_URL, &ADCM_LOGIN, &ADCM_PASSWORD)
		if rs.Primary.ID == "" {
			return fmt.Errorf("failed to create adcmClient: %s", err)
		}
		ui64, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to cinvert id: %s", err)
		}
		bundle, err := client.GetBundle(adcmClient.BundleSearch{Identifier: adcmClient.Identifier{ID: ui64}})
		if err != nil {
			return fmt.Errorf("failed to find bundle: %s", err)
		}

		if bundle.Edition != "community" {
			return fmt.Errorf("unexpected bundle edition")
		}

		return nil
	}
}

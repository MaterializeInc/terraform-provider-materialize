package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceSSOConfig_basic(t *testing.T) {
	resourceName := "data.materialize_sso_config.test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSSOConfigBasicConfig(true, true, "https://example.com/sso", "test-certificate", "saml", "client-id", "secret"),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceSSOConfigExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "sso_configs.0.enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "sso_configs.0.sso_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "sso_configs.0.public_certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "sso_configs.0.sign_request"),
					resource.TestCheckResourceAttrSet(resourceName, "sso_configs.0.type"),
					resource.TestCheckResourceAttrSet(resourceName, "sso_configs.0.oidc_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "sso_configs.0.oidc_secret"),
					resource.TestCheckResourceAttr(resourceName, "sso_configs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "sso_configs.0.sso_endpoint", "https://example.com/sso"),
					resource.TestCheckResourceAttr(resourceName, "sso_configs.0.public_certificate", "dGVzdC1jZXJ0aWZpY2F0ZQ=="),
					resource.TestCheckResourceAttr(resourceName, "sso_configs.0.sign_request", "true"),
					resource.TestCheckResourceAttr(resourceName, "sso_configs.0.type", "saml"),
					resource.TestCheckResourceAttr(resourceName, "sso_configs.0.oidc_client_id", "client-id"),
					resource.TestCheckResourceAttr(resourceName, "sso_configs.0.oidc_secret", "secret"),
				),
			},
		},
	})
}

func testAccDataSourceSSOConfigBasicConfig(enabled, signRequest bool, ssoEndpoint, publicCertificate, ssoType, oidcClientId, oidcSecret string) string {
	return fmt.Sprintf(`
	resource "materialize_sso_config" "example" {
		enabled             = %t
		sign_request        = %t
		sso_endpoint        = "%s"
		public_certificate  = "%s"
		type                = "%s"
		oidc_client_id      = "%s"
		oidc_secret         = "%s"
	}

	data "materialize_sso_config" "test" {
		depends_on = [materialize_sso_config.example]
	}
	`, enabled, signRequest, ssoEndpoint, publicCertificate, ssoType, oidcClientId, oidcSecret)
}

func testAccDataSourceSSOConfigExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSO Config ID is set")
		}

		return nil
	}
}

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSSOConfiguration_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOConfigBasicConfig(true, true, "https://example.com/sso", "test-certificate", "saml", "client-id", "secret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSOConfigExists("materialize_sso_config.example"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "enabled", "true"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "sso_endpoint", "https://example.com/sso"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "public_certificate", "test-certificate"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "sign_request", "true"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "type", "saml"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "oidc_client_id", "client-id"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "oidc_secret", "secret"),
				),
			},
		},
	})
}

func TestAccSSOConfiguration_update(t *testing.T) {
	initialSSOEndpoint := "https://example.com/sso"
	updatedSSOEndpoint := "https://updated.example.com/sso"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOConfigBasicConfig(true, true, initialSSOEndpoint, "test-certificate", "saml", "client-id", "secret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSOConfigExists("materialize_sso_config.example"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "sso_endpoint", initialSSOEndpoint),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "public_certificate", "test-certificate"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "sign_request", "true"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "type", "saml"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "oidc_client_id", "client-id"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "oidc_secret", "secret"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "enabled", "true"),
				),
			},
			{
				Config: testAccSSOConfigBasicConfig(true, true, updatedSSOEndpoint, "test-certificate2", "saml", "client-id", "secret2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSOConfigExists("materialize_sso_config.example"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "sso_endpoint", updatedSSOEndpoint),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "public_certificate", "test-certificate2"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "sign_request", "true"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "type", "saml"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "oidc_client_id", "client-id"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "oidc_secret", "secret2"),
					resource.TestCheckResourceAttr("materialize_sso_config.example", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccSSOConfiguration_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOConfigBasicConfig(true, true, "https://example.com/sso", "test-certificate", "saml", "client-id", "secret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSOConfigExists("materialize_sso_config.example"),
					testAccCheckSSOConfigDisappears("materialize_sso_config.example"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSSOConfigBasicConfig(enabled, signRequest bool, ssoEndpoint, publicCertificate, ssoType, oidcClientId, oidcSecret string) string {
	// Return a Terraform configuration for an example SSO Configuration
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
`, enabled, signRequest, ssoEndpoint, publicCertificate, ssoType, oidcClientId, oidcSecret)
}

func testAccCheckSSOConfigExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSO Configuration ID is set")
		}

		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg

		configurations, err := frontegg.FetchSSOConfigurations(context.Background(), client)
		if err != nil {
			return err
		}

		var foundConfig *frontegg.SSOConfig
		for _, config := range configurations {
			if config.Id == rs.Primary.ID {
				foundConfig = &config
				break
			}
		}

		if foundConfig == nil {
			return fmt.Errorf("SSO Configuration not found")
		}

		return nil
	}
}

func testAccCheckSSOConfigNotExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg

		configurations, err := frontegg.FetchSSOConfigurations(context.Background(), client)
		if err != nil {
			return fmt.Errorf("Error fetching SSO configurations: %s", err)
		}

		for _, config := range configurations {
			if config.Id == rs.Primary.ID {
				return fmt.Errorf("SSO configuration %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckSSOConfigDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg

		err := frontegg.DeleteSSOConfiguration(context.Background(), client, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error deleting SSO configuration: %s", err)
		}

		return nil
	}
}

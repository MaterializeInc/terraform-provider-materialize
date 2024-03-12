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

func TestAccSSODomain_basic(t *testing.T) {
	resourceName := "materialize_sso_domain.example"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSODomainConfig("example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSODomainExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "domain", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "validated", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func TestAccSSODomain_update(t *testing.T) {
	resourceName := "materialize_sso_domain.example"
	initialDomain := "example.com"
	updatedDomain := "updated-example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSODomainConfig(initialDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSODomainExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "domain", initialDomain),
					resource.TestCheckResourceAttr(resourceName, "validated", "false"),
				),
			},
			{
				Config: testAccSSODomainConfig(updatedDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSODomainExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "domain", updatedDomain),
					resource.TestCheckResourceAttr(resourceName, "validated", "false"),
				),
			},
		},
	})
}

func TestAccSSODomain_disappears(t *testing.T) {
	resourceName := "materialize_sso_domain.example"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSODomainConfig("example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSODomainExists(resourceName),
					testAccCheckSSODomainDisappears(resourceName),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSSODomainConfig(domain string) string {
	return fmt.Sprintf(`
resource "materialize_sso_config" "example" {
  enabled             = false
  sign_request        = false
  sso_endpoint        = "https://example.com/sso"
  public_certificate  = "test-certificate"
  type                = "saml"
}
resource "materialize_sso_domain" "example" {
  sso_config_id = materialize_sso_config.example.id
  domain        = "%s"
  depends_on    = [materialize_sso_config.example]
}
`, domain)
}

func testAccCheckSSODomainExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSO Domain ID is set")
		}

		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg

		domain, err := frontegg.FetchSSODomain(context.Background(), client, rs.Primary.Attributes["sso_config_id"], rs.Primary.Attributes["domain"])
		if err != nil {
			return err
		}

		if domain == nil || domain.ID != rs.Primary.ID {
			return fmt.Errorf("SSO Domain not found")
		}

		return nil
	}
}

func testAccCheckSSODomainDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg

		err := frontegg.DeleteSSODomain(context.Background(), client, rs.Primary.Attributes["sso_config_id"], rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error deleting SSO domain: %s", err)
		}

		return nil
	}
}

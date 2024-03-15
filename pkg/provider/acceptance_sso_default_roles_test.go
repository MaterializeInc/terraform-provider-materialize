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

func TestAccSSODefaultRoles_basic(t *testing.T) {
	resourceName := "materialize_sso_default_roles.example"
	roleIDs := []string{"Admin", "Member"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSODefaultRolesConfig(roleIDs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSODefaultRolesExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "roles.0", roleIDs[0]),
					resource.TestCheckResourceAttr(resourceName, "roles.1", roleIDs[1]),
					resource.TestCheckResourceAttrSet(resourceName, "sso_config_id"),
				),
			},
		},
	})
}

func TestAccSSODefaultRoles_update(t *testing.T) {
	resourceName := "materialize_sso_default_roles.example"
	initialRoleIDs := []string{"Admin", "Member"}
	updatedRoleIDs := []string{"Admin"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSODefaultRolesConfig(initialRoleIDs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSODefaultRolesExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "2"),
				),
			},
			{
				Config: testAccSSODefaultRolesConfig(updatedRoleIDs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSODefaultRolesExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "roles.0", updatedRoleIDs[0]),
				),
			},
		},
	})
}

func TestAccSSODefaultRoles_disappears(t *testing.T) {
	resourceName := "materialize_sso_default_roles.example"
	roleIDs := []string{"Admin", "Member"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSODefaultRolesConfig(roleIDs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSODefaultRolesExists(resourceName),
					testAccCheckSSODefaultRolesDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSSODefaultRolesConfig(roleIDs []string) string {
	rolesStr := ""
	for _, id := range roleIDs {
		rolesStr += fmt.Sprintf("\"%s\",", id)
	}

	return fmt.Sprintf(`
	resource "materialize_sso_config" "example" {
		enabled             = false
		sign_request        = false
		sso_endpoint        = "https://example.com/sso"
		public_certificate  = "test-certificate"
		type                = "saml"
	  }
	resource "materialize_sso_default_roles" "example" {
		sso_config_id = materialize_sso_config.example.id
		roles         = [%s]
	}
`, rolesStr)
}

func testAccCheckSSODefaultRolesExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSO Default Roles ID is set")
		}

		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg

		roleIDs, err := frontegg.GetSSODefaultRoles(context.Background(), client, rs.Primary.Attributes["sso_config_id"])
		if err != nil {
			return err
		}

		if len(roleIDs) == 0 {
			return fmt.Errorf("SSO Default Roles not found")
		}

		return nil
	}
}

func testAccCheckSSODefaultRolesDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg

		err := frontegg.ClearSSODefaultRoles(context.Background(), client, rs.Primary.Attributes["sso_config_id"])
		if err != nil {
			return fmt.Errorf("Error clearing SSO default roles: %s", err)
		}

		return nil
	}
}

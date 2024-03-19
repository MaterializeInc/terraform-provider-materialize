package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSSORoleGroupMapping_basic(t *testing.T) {
	resourceName := "materialize_sso_group_mapping.example"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSORoleGroupMappingConfig("group1", []string{"Admin", "Member"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSORoleGroupMappingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "group", "group1"),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func TestAccSSORoleGroupMapping_update(t *testing.T) {
	resourceName := "materialize_sso_group_mapping.example"
	initialGroup := "Member"
	updatedGroup := "Admin"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSORoleGroupMappingConfig(initialGroup, []string{"Member"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSORoleGroupMappingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "group", initialGroup),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "1"),
				),
			},
			{
				Config: testAccSSORoleGroupMappingConfig(updatedGroup, []string{"Member", "Admin"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSORoleGroupMappingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "group", updatedGroup),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "2"),
				),
			},
		},
	})
}

func TestAccSSORoleGroupMapping_disappears(t *testing.T) {
	resourceName := "materialize_sso_group_mapping.example"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSORoleGroupMappingConfig("group1", []string{"Admin", "Member"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSORoleGroupMappingExists(resourceName),
					testAccCheckSSORoleGroupMappingDisappears(resourceName),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSSORoleGroupMappingConfig(group string, roles []string) string {
	rolesStr := fmt.Sprintf("[\"%s\"]", strings.Join(roles, "\",\""))

	return fmt.Sprintf(`
	resource "materialize_sso_config" "example" {
		enabled             = false
		sign_request        = false
		sso_endpoint        = "https://example.com/sso"
		public_certificate  = "test-certificate"
		type                = "saml"
	}
	resource "materialize_sso_group_mapping" "example" {
		sso_config_id = materialize_sso_config.example.id
		group         = "%s"
		roles         = %s
	}
	`, group, rolesStr)
}

func testAccCheckSSORoleGroupMappingExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSO Role Group Mapping ID is set")
		}

		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg

		groupMapping, err := frontegg.FetchSSOGroupMapping(context.Background(), client, rs.Primary.Attributes["sso_config_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		if groupMapping == nil || groupMapping.ID != rs.Primary.ID {
			return fmt.Errorf("SSO Role Group Mapping not found")
		}

		return nil
	}
}

func testAccCheckSSORoleGroupMappingDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg

		err := frontegg.DeleteSSOGroupMapping(context.Background(), client, rs.Primary.Attributes["sso_config_id"], rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error deleting SSO Role Group Mapping: %s", err)
		}

		return nil
	}
}

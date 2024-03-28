package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSCIM2GroupRoles_basic(t *testing.T) {
	groupName := fmt.Sprintf("test-group-%d", time.Now().UnixNano())
	roleNames := []string{"Admin", "Member"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSCIM2GroupRolesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSCIM2GroupRolesConfig(groupName, roleNames),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2GroupRolesExists("materialize_scim_group_roles.scim_group_roles_example"),
					resource.TestCheckResourceAttr("materialize_scim_group.scim_group_example", "name", groupName),
					resource.TestCheckResourceAttr("materialize_scim_group_roles.scim_group_roles_example", "roles.#", "2"),
				),
			},
		},
	})
}

func TestAccSCIM2GroupRoles_update(t *testing.T) {
	groupName := fmt.Sprintf("test-group-%d", time.Now().UnixNano())
	initialRoles := []string{"Admin"}
	updatedRoles := []string{"Member"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSCIM2GroupRolesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSCIM2GroupRolesConfig(groupName, initialRoles),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2GroupRolesExists("materialize_scim_group_roles.scim_group_roles_example"),
					resource.TestCheckResourceAttr("materialize_scim_group_roles.scim_group_roles_example", "roles.#", "1"),
					resource.TestCheckResourceAttr("materialize_scim_group_roles.scim_group_roles_example", "roles.0", "Admin"),
				),
			},
			{
				Config: testAccSCIM2GroupRolesConfig(groupName, updatedRoles),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2GroupRolesExists("materialize_scim_group_roles.scim_group_roles_example"),
					resource.TestCheckResourceAttr("materialize_scim_group_roles.scim_group_roles_example", "roles.#", "1"),
					resource.TestCheckResourceAttr("materialize_scim_group_roles.scim_group_roles_example", "roles.0", "Member"),
				),
			},
		},
	})
}

func TestAccSCIM2GroupRoles_disappears(t *testing.T) {
	groupName := fmt.Sprintf("test-group-%d", time.Now().UnixNano())
	roleNames := []string{"Admin", "Member"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSCIM2GroupRolesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSCIM2GroupRolesConfig(groupName, roleNames),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2GroupRolesExists("materialize_scim_group_roles.scim_group_roles_example"),
					testAccCheckSCIM2GroupRolesDisappears("materialize_scim_group_roles.scim_group_roles_example"),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSCIM2GroupRolesConfig(groupName string, roles []string) string {
	// Convert roles slice to a Terraform set syntax
	rolesStr := fmt.Sprintf(`["%s"]`, strings.Join(roles, `", "`))

	return fmt.Sprintf(`
resource "materialize_scim_group" "scim_group_example" {
  name        = "%s"
  description = "A test SCIM group for roles"
}

resource "materialize_scim_group_roles" "scim_group_roles_example" {
  group_id = materialize_scim_group.scim_group_example.id
  roles    = %s
}
`, groupName, rolesStr)
}

func testAccCheckSCIM2GroupRolesExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SCIM Group Roles ID is set")
		}

		return nil
	}
}

func testAccCheckSCIM2GroupRolesDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg

		groupID := rs.Primary.ID
		err := frontegg.DeleteSCIMGroup(context.Background(), client, groupID)
		if err != nil {
			return fmt.Errorf("error deleting SCIM group outside of Terraform: %s", err)
		}

		return nil
	}
}

func testAccCheckSCIM2GroupRolesDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta()
	providerMeta, _ := utils.GetProviderMeta(meta)
	client := providerMeta.Frontegg

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "materialize_scim_group_roles" {
			continue
		}

		groupID := rs.Primary.ID
		group, err := frontegg.GetSCIMGroupByID(context.Background(), client, groupID)
		if err == nil && group != nil {
			return fmt.Errorf("SCIM group roles for group %s still exists", groupID)
		}
	}

	return nil
}

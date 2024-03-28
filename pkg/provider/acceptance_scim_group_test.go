package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSCIM2Group_basic(t *testing.T) {
	groupName := fmt.Sprintf("test-group-%d", time.Now().UnixNano())
	groupDescription := "A test SCIM group"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSCIM2GroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSCIM2GroupConfig(groupName, groupDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2GroupExists("materialize_scim_group.scim_group_example"),
					resource.TestCheckResourceAttr("materialize_scim_group.scim_group_example", "name", groupName),
					resource.TestCheckResourceAttr("materialize_scim_group.scim_group_example", "description", groupDescription),
				),
			},
		},
	})
}

func TestAccSCIM2Group_update(t *testing.T) {
	initialGroupName := fmt.Sprintf("test-group-%d", time.Now().UnixNano())
	initialGroupDescription := "A test SCIM group"
	updatedGroupName := initialGroupName + "-updated"
	updatedGroupDescription := initialGroupDescription + " - updated"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSCIM2GroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSCIM2GroupConfig(initialGroupName, initialGroupDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2GroupExists("materialize_scim_group.scim_group_example"),
					resource.TestCheckResourceAttr("materialize_scim_group.scim_group_example", "name", initialGroupName),
					resource.TestCheckResourceAttr("materialize_scim_group.scim_group_example", "description", initialGroupDescription),
				),
			},
			{
				Config: testAccSCIM2GroupConfig(updatedGroupName, updatedGroupDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2GroupExists("materialize_scim_group.scim_group_example"),
					resource.TestCheckResourceAttr("materialize_scim_group.scim_group_example", "name", updatedGroupName),
					resource.TestCheckResourceAttr("materialize_scim_group.scim_group_example", "description", updatedGroupDescription),
				),
			},
		},
	})
}

func TestAccSCIM2Group_disappears(t *testing.T) {
	groupName := fmt.Sprintf("test-group-%d", time.Now().UnixNano())
	groupDescription := "A test SCIM group"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSCIM2GroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSCIM2GroupConfig(groupName, groupDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2GroupExists("materialize_scim_group.scim_group_example"),
					testAccCheckSCIM2GroupDisappears("materialize_scim_group.scim_group_example"),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSCIM2GroupConfig(name, description string) string {
	return fmt.Sprintf(`
resource "materialize_scim_group" "scim_group_example" {
  name        = "%s"
  description = "%s"
}
`, name, description)
}

func testAccCheckSCIM2GroupExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SCIM Group ID is set")
		}

		return nil
	}
}

func testAccCheckSCIM2GroupDisappears(resourceName string) resource.TestCheckFunc {
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

func testAccCheckSCIM2GroupDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta()
	providerMeta, _ := utils.GetProviderMeta(meta)
	client := providerMeta.Frontegg

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "materialize_scim_group" {
			continue
		}

		groupID := rs.Primary.ID
		_, err := frontegg.GetSCIMGroupByID(context.Background(), client, groupID)
		if err == nil {
			return fmt.Errorf("SCIM group %s still exists", groupID)
		}
	}

	return nil
}

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

func TestAccSCIM2GroupUsers_basic(t *testing.T) {
	resourceName := "materialize_scim_group_users.scim_group_users_example"
	groupName := fmt.Sprintf("test-group-%d", time.Now().UnixNano())
	userEmail := "test@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSCIM2GroupUsersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSCIM2GroupUsersConfig(groupName, userEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2GroupUsersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "users.#", "1"),
				),
			},
		},
	})
}

func TestAccSCIM2GroupUsers_update(t *testing.T) {
	resourceName := "materialize_scim_group_users.scim_group_users_example"
	groupName := fmt.Sprintf("test-group-%d", time.Now().UnixNano())
	initialUserEmail := "test1@example.com"
	updatedUserEmail := "test2@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSCIM2GroupUsersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSCIM2GroupUsersConfig(groupName, initialUserEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2GroupUsersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "users.#", "1"),
				),
			},
			{
				Config: testAccSCIM2GroupUsersConfig(groupName, updatedUserEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2GroupUsersExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "users.#", "1"),
				),
			},
		},
	})
}

func TestAccSCIM2GroupUsers_disappears(t *testing.T) {
	resourceName := "materialize_scim_group_users.scim_group_users_example"
	groupName := fmt.Sprintf("test-group-%d", time.Now().UnixNano())
	userEmail := "test@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSCIM2GroupUsersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSCIM2GroupUsersConfig(groupName, userEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2GroupUsersExists(resourceName),
					testAccCheckSCIM2GroupUsersDisappears(resourceName),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSCIM2GroupUsersConfig(groupName, userEmail string) string {
	return fmt.Sprintf(`
resource "materialize_scim_group" "scim_group_example" {
  name        = "%s"
  description = "Test group for SCIM users"
}

resource "materialize_user" "example_user" {
  email = "%s"
  roles = ["Member"]
}

resource "materialize_scim_group_users" "scim_group_users_example" {
  group_id = materialize_scim_group.scim_group_example.id
  users    = [materialize_user.example_user.id]
}
`, groupName, userEmail)
}

func testAccCheckSCIM2GroupUsersExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SCIM Group Users ID is set")
		}

		return nil
	}
}

func testAccCheckSCIM2GroupUsersDisappears(resourceName string) resource.TestCheckFunc {
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

func testAccCheckSCIM2GroupUsersDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta()
	providerMeta, _ := utils.GetProviderMeta(meta)
	client := providerMeta.Frontegg

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "materialize_scim_group_users" {
			continue
		}

		groupID := rs.Primary.ID
		group, err := frontegg.GetSCIMGroupByID(context.Background(), client, groupID)
		// If the group doesn't exist, the deletion was successful
		if err != nil {
			continue
		}

		// If the group exists, check if it still has users
		if len(group.Users) != 0 {
			return fmt.Errorf("SCIM group %s still has users after deletion", groupID)
		}
	}

	return nil
}

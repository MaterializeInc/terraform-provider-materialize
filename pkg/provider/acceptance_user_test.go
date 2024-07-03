package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUser_basic(t *testing.T) {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(email, false, "Member"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("materialize_user.example_user", email),
					resource.TestCheckResourceAttr("materialize_user.example_user", "email", email),
					resource.TestCheckResourceAttr("materialize_user.example_user", "send_activation_email", "false"),
					resource.TestCheckResourceAttr("materialize_user.example_user", "roles.0", "Member"),
					resource.TestCheckResourceAttr("materialize_user.example_user", "verified", "false"),
					// Data source tests
					resource.TestCheckResourceAttrPair("data.materialize_user.user_data", "id", "materialize_user.example_user", "id"),
					resource.TestCheckResourceAttr("data.materialize_user.user_data", "email", email),
					resource.TestCheckResourceAttr("data.materialize_user.user_data", "verified", "false"),
				),
			},
		},
	})
}

func TestAccUser_disappears(t *testing.T) {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(email, true, "Member"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("materialize_user.example_user", email),
					resource.TestCheckResourceAttr("materialize_user.example_user", "email", email),
					resource.TestCheckResourceAttr("materialize_user.example_user", "roles.0", "Member"),
					testAccCheckUserDestroy,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccUser_updateRole(t *testing.T) {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(email, true, "Member"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("materialize_user.example_user", email),
					resource.TestCheckResourceAttr("materialize_user.example_user", "email", email),
					resource.TestCheckResourceAttr("materialize_user.example_user", "roles.0", "Member"),
					resource.TestCheckResourceAttr("materialize_user.example_user", "verified", "false"),
				),
			},
			{
				Config: testAccUserConfig(email, true, "Admin"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("materialize_user.example_user", email),
					resource.TestCheckResourceAttr("materialize_user.example_user", "roles.0", "Admin"),
				),
			},
		},
	})
}

func TestAccUserDataSource_nonExistent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserDataSourceConfigNonExistent(),
				ExpectError: regexp.MustCompile(`No user found with email:`),
			},
		},
	})
}

func testAccUserConfig(email string, sendActivationEmail bool, role string) string {
	return fmt.Sprintf(`
resource "materialize_user" "example_user" {
  email = "%s"
  send_activation_email = %v
  roles = ["%s"]
}
data "materialize_user" "user_data" {
  depends_on = [materialize_user.example_user]
  email = materialize_user.example_user.email
}
`, email, sendActivationEmail, role)
}

func testAccUserDataSourceConfigNonExistent() string {
	return `
data "materialize_user" "nonexistent" {
  email = "nonexistent@example.com"
}
`
}

func testAccCheckUserExists(resourceName, email string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		userID := rs.Primary.ID
		_, err := frontegg.ReadUser(context.Background(), client, userID)
		if err != nil {
			return fmt.Errorf("Error fetching user with ID [%s]: %s", userID, err)
		}

		return nil
	}
}

func testAccCheckUserDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "materialize_user" {
			continue
		}

		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg

		_, err := frontegg.ReadUser(context.Background(), client, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("User with ID [%s] still exists", rs.Primary.ID)
		}
	}

	return nil
}

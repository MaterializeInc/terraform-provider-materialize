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

func TestAccAppPassword_basic(t *testing.T) {
	resourceName := "materialize_app_password.test"
	name := "test-name"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAppPasswordConfigBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-name"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
					resource.TestCheckResourceAttrSet(resourceName, "password"),
				),
			},
		},
	})
}

// Disappears test
func TestAccAppPassword_disappears(t *testing.T) {
	name := "test-name"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAppPasswordConfigBasic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppPasswordExists("materialize_app_password.test"),
					resource.TestCheckResourceAttr("materialize_app_password.test", "test-name", name),
					resource.TestCheckResourceAttrSet("materialize_app_password.test", "created_at"),
					testAccCheckAppPasswordDestroy,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAppPasswordConfigBasic(name string) string {
	// Return a basic Terraform configuration for your resource
	return fmt.Sprintf(`
		resource "materialize_app_password" "test" {
			name = "%[1]s"
		}
	`, name)
}

func testAccCheckAppPasswordExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		appPassword, err := frontegg.ListAppPasswords(context.Background(), client)
		if err != nil {
			return fmt.Errorf("Error fetching app password with resource ID [%s]: %s", rs.Primary.ID, err)
		}

		if appPassword == nil {
			return fmt.Errorf("App password with resource ID [%s] does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAppPasswordDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "materialize_app_password" {
			continue
		}

		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg

		appPassword, err := frontegg.ListAppPasswords(context.Background(), client)
		if err == nil {
			return fmt.Errorf("App password with ID [%s] still exists", rs.Primary.ID)
		}

		if appPassword != nil {
			return fmt.Errorf("App password with ID [%s] still exists", rs.Primary.ID)
		}
	}

	return nil
}

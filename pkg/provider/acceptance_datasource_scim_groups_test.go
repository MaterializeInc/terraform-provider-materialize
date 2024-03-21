package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceSCIMGroups_basic(t *testing.T) {
	resourceName := "data.materialize_scim_groups.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSCIMGroupsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceSCIMGroupsExists(resourceName),
					// TODO: Add more checks once SCIM Groups resource is implemented
					resource.TestCheckResourceAttr(resourceName, "groups.#", "0"),
				),
			},
		},
	})
}

func testAccDataSourceSCIMGroupsConfig() string {
	return fmt.Sprintf(`
data "materialize_scim_groups" "test" {}
`)
}

func testAccDataSourceSCIMGroupsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SCIM Groups data source ID is set")
		}

		return nil
	}
}

package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceSCIM2Configurations_basic(t *testing.T) {
	resourceName := "data.materialize_scim_configs.test"
	resourceType := "materialize_scim_config"
	resourceLabel := "example"
	source := "okta"
	connectionName := fmt.Sprintf("test-conn-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSCIM2ConfigurationsConfig(source, connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2ConfigurationExists(fmt.Sprintf("%s.%s", resourceType, resourceLabel), source, connectionName),
					testAccDataSourceSCIM2ConfigurationsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.source", source),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.connection_name", connectionName),
					resource.TestCheckResourceAttrSet(resourceName, "configurations.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "configurations.0.tenant_id"),
					resource.TestCheckResourceAttrSet(resourceName, "configurations.0.sync_to_user_management"),
				),
			},
		},
	})
}

func testAccDataSourceSCIM2ConfigurationsConfig(source, connectionName string) string {
	return fmt.Sprintf(`
resource "%s" "%s" {
  source = "%s"
  connection_name = "%s"
}

data "%s" "%s" {
  depends_on = ["%s.%s"]
}
`, "materialize_scim_config", "example", source, connectionName, "materialize_scim_configs", "test", "materialize_scim_config", "example")
}

func testAccDataSourceSCIM2ConfigurationsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SCIM 2.0 Configurations data source ID is set")
		}

		return nil
	}
}

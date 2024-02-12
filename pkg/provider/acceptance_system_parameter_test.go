package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSystemParameter_basic(t *testing.T) {
	resourceName := "materialize_system_parameter.test"
	paramName := "max_connections"
	paramValue := "100"
	paramValueUpdated := "200"
	randomID := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSystemParameterConfig(randomID, paramName, paramValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", paramName),
					resource.TestCheckResourceAttr(resourceName, "value", paramValue),
				),
			},
			{
				Config: testAccSystemParameterConfig(randomID, paramName, paramValueUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", paramName),
					resource.TestCheckResourceAttr(resourceName, "value", paramValueUpdated),
				),
			},
		},
	})
}

func testAccSystemParameterConfig(id, paramName, paramValue string) string {
	return fmt.Sprintf(`
resource "materialize_system_parameter" "test" {
	name  = "%s"
	value = "%s"
}
`, paramName, paramValue)
}

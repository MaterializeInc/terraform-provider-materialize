package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceCurrentDatabase_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceCurrentDatabase(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.materialize_current_database.current", "name", "materialize"),
				),
			},
		},
	})
}

func testAccDatasourceCurrentDatabase() string {
	return `data "materialize_current_database" "current" {}`
}

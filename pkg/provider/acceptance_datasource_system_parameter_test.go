package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceSystemParameters_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceSystemParameters(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.materialize_system_parameter.all", "parameters.#"),
					resource.TestCheckResourceAttrSet("data.materialize_system_parameter.all", "parameters.0.name"),
					resource.TestCheckResourceAttrSet("data.materialize_system_parameter.all", "parameters.0.setting"),
					resource.TestCheckResourceAttrSet("data.materialize_system_parameter.all", "parameters.0.description"),
				),
			},
		},
	})
}

func testAccDatasourceSystemParameters() string {
	return `
data "materialize_system_parameter" "all" {}
`
}

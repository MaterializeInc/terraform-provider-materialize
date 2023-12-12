package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceEgressIp_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceEgressIp(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.materialize_egress_ips.current", "egress_ips.#", "0"),
				),
			},
		},
	})
}

func testAccDatasourceEgressIp() string {
	return `data "materialize_egress_ips" "current" {}`
}

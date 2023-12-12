package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceCurrentCluster_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceCurrentCluster(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.materialize_current_cluster.current", "name", "mz_system"),
				),
			},
		},
	})
}

func testAccDatasourceCurrentCluster() string {
	return `data "materialize_current_cluster" "current" {}`
}

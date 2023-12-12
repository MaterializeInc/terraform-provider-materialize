package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceCluster_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceCluster(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					// Cannot ensure the exact number of objects with parallel tests
					// Ensuring minimum
					resource.TestMatchResourceAttr("data.materialize_cluster.test_all", "clusters.#", regexp.MustCompile("([4-9]|\\d{2,})")),
				),
			},
		},
	})
}

func testAccDatasourceCluster(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_cluster" "a" {
		name         = "%[1]s_a"
		size         = "3xsmall"
	}

	resource "materialize_cluster" "b" {
		name         = "%[1]s_b"
		size         = "2xsmall"
	}

	resource "materialize_cluster" "c" {
		name         = "%[1]s_c"
		size         = "2xsmall"
	}

	resource "materialize_cluster" "d" {
		name         = "%[1]s_d"
		size         = "3xsmall"
	}


	data "materialize_cluster" "test_all" {
		depends_on    = [
			materialize_cluster.a,
			materialize_cluster.b,
			materialize_cluster.c,
			materialize_cluster.d,
		]
	}
	`, nameSpace)
}

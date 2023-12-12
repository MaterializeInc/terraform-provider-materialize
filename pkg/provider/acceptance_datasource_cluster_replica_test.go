package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceClusterReplica_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceClusterReplica(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					// Cannot ensure the exact number of objects with parallel tests
					// Ensuring minimum
					resource.TestMatchResourceAttr("data.materialize_cluster_replica.test_all", "cluster_replicas.#", regexp.MustCompile("([5-9]|\\d{2,})")),
				),
			},
		},
	})
}

func testAccDatasourceClusterReplica(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_cluster" "test" {
		name    = "%[1]s"
	}

	resource "materialize_cluster" "test_2" {
		name    = "%[1]s_2"
	}

	resource "materialize_cluster_replica" "a" {
		name         = "%[1]s_a"
		cluster_name = materialize_cluster.test.name
		size         = "3xsmall"
	}

	resource "materialize_cluster_replica" "b" {
		name         = "%[1]s_b"
		cluster_name = materialize_cluster.test.name
		size         = "2xsmall"
	}

	resource "materialize_cluster_replica" "c" {
		name         = "%[1]s_c"
		cluster_name = materialize_cluster.test_2.name
		size         = "2xsmall"
	}

	resource "materialize_cluster_replica" "d" {
		name         = "%[1]s_d"
		cluster_name = materialize_cluster.test_2.name
		size         = "3xsmall"
	}

	resource "materialize_cluster_replica" "e" {
		name         = "%[1]s_e"
		cluster_name = materialize_cluster.test_2.name
		size         = "2xsmall"
	}


	data "materialize_cluster_replica" "test_all" {
		depends_on    = [
			materialize_cluster_replica.a,
			materialize_cluster_replica.b,
			materialize_cluster_replica.c,
			materialize_cluster_replica.d,
			materialize_cluster_replica.e,
		]
	}
	`, nameSpace)
}

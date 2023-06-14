package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jmoiron/sqlx"
)

func TestAccClusterReplica_basic(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	replicaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterReplicaResource(clusterName, replicaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterReplicaExists("materialize_cluster_replica.test"),
					resource.TestCheckResourceAttr("materialize_cluster_replica.test", "cluster_name", clusterName),
					resource.TestCheckResourceAttr("materialize_cluster_replica.test", "name", replicaName),
					resource.TestCheckResourceAttr("materialize_cluster_replica.test", "size", "1"),
					resource.TestCheckResourceAttr("materialize_cluster_replica.test", "introspection_interval", "1s"),
					resource.TestCheckResourceAttr("materialize_cluster_replica.test", "introspection_debugging", "false"),
					resource.TestCheckNoResourceAttr("materialize_cluster_replica.test", "idle_arrangement_merge_effort"),
				),
			},
		},
	})
}

func TestAccClusterReplica_disappears(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	replicaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterReplicaResource(clusterName, replicaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterReplicaExists("materialize_cluster_replica.test"),
					testAccCheckClusterReplicaDisappears(clusterName, replicaName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccClusterReplicaResource(clusterName, replicaName string) string {
	return fmt.Sprintf(`
resource "materialize_cluster" "replica_cluster" {
	name = "%[1]s"
}

resource "materialize_cluster_replica" "test" {
	cluster_name = materialize_cluster.replica_cluster.name
	name = "%[2]s"
	size = "1"
}
`, clusterName, replicaName)
}

func testAccCheckClusterReplicaExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("cluster replica not found: %s", name)
		}
		_, err := materialize.ScanClusterReplica(db, r.Primary.ID)
		return err
	}
}

func testAccCheckClusterReplicaDisappears(clusterName, replicaName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP CLUSTER REPLICA "%s"."%s";`, clusterName, replicaName))
		return err
	}
}

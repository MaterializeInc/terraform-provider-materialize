package provider

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jmoiron/sqlx"
)

func TestAccCluster_basic(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	cluster2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	clusterSize := "2-2"
	clusterReplicationFactor := "1"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterResource(roleName, clusterName, cluster2Name, roleName, clusterSize, clusterReplicationFactor),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "name", clusterName),
					resource.TestCheckResourceAttr("materialize_cluster.test", "ownership_role", "mz_system"),
					testAccCheckClusterExists("materialize_cluster.test_role"),
					resource.TestCheckResourceAttr("materialize_cluster.test_role", "name", cluster2Name),
					resource.TestCheckResourceAttr("materialize_cluster.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_cluster.test", "replication_factor", "0"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "size", ""),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "name", clusterName+"_managed"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "replication_factor", "1"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "size", "2-2"),
				),
			},
		},
	})
}

func TestAccCluster_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	oldClusterName := fmt.Sprintf("old_%s", slug)
	newClusterName := fmt.Sprintf("new_%s", slug)
	cluster2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	newClusterSize := "2-2"
	newClusterReplicationFactor := "1"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterResource(roleName, oldClusterName, cluster2Name, "mz_system", "2-1", "2"),
			},
			{
				Config: testAccClusterResource(roleName, newClusterName, cluster2Name, roleName, newClusterSize, newClusterReplicationFactor),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "name", newClusterName),
					resource.TestCheckResourceAttr("materialize_cluster.test", "ownership_role", "mz_system"),
					testAccCheckClusterExists("materialize_cluster.test_role"),
					resource.TestCheckResourceAttr("materialize_cluster.test_role", "name", cluster2Name),
					resource.TestCheckResourceAttr("materialize_cluster.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_cluster.test", "replication_factor", "0"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "size", ""),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "name", newClusterName+"_managed"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "replication_factor", "1"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "size", "2-2"),
				),
			},
		},
	})
}

func TestAccCluster_disappears(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	cluster2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	clusterSize := "2-2"
	clusterReplicationFactor := "1"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllClusterDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterResource(roleName, clusterName, cluster2Name, roleName, clusterSize, clusterReplicationFactor),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					testAccCheckObjectDisappears(materialize.ObjectSchemaStruct{ObjectType: "CLUSTER", Name: clusterName}),
					testAccCheckClusterExists("materialize_cluster.test_managed_cluster"),
					testAccCheckObjectDisappears(materialize.ObjectSchemaStruct{ObjectType: "CLUSTER", Name: clusterName + "_managed"}),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccClusterResource(roleName, cluster1Name, cluster2Name, cluster2Owner, clusterSize, clusterReplicationFactor string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_cluster" "test" {
	name = "%[2]s"
}

resource "materialize_cluster" "test_role" {
	name = "%[3]s"
	ownership_role = "%[4]s"

	depends_on = [materialize_role.test]
}

resource "materialize_cluster" "test_managed_cluster" {
	name               = "%[2]s_managed"
	size               = "%[5]s"
	replication_factor = %[6]s
}

`, roleName, cluster1Name, cluster2Name, cluster2Owner, clusterSize, clusterReplicationFactor)
}

func testAccCheckClusterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("cluster not found: %s", name)
		}
		_, err := materialize.ScanCluster(db, r.Primary.ID)
		return err
	}
}

func testAccCheckAllClusterDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_cluster" {
			continue
		}

		_, err := materialize.ScanCluster(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("Cluster %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

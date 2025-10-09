package provider

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCluster_basic(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	cluster2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterResource(roleName, clusterName, cluster2Name, roleName, "3xsmall", "1", "1s", "true", "true", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					resource.TestMatchResourceAttr("materialize_cluster.test", "id", terraformObjectTypeIdRegex),
					resource.TestCheckResourceAttr("materialize_cluster.test", "name", clusterName),
					resource.TestCheckResourceAttr("materialize_cluster.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "size", ""),
					testAccCheckClusterExists("materialize_cluster.test_role"),
					resource.TestCheckResourceAttr("materialize_cluster.test_role", "name", cluster2Name),
					resource.TestCheckResourceAttr("materialize_cluster.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_cluster.test_role", "size", ""),
					testAccCheckClusterExists("materialize_cluster.test_managed_cluster"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "name", clusterName+"_managed"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "size", "3xsmall"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "replication_factor", "1"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "introspection_interval", "1s"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "introspection_debugging", "true"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "disk", "true"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "availability_zones.#", "2"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "comment", "Comment"),
				),
			},
			{
				ResourceName:            "materialize_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"introspection_debugging", "introspection_interval", "identify_by_name"},
			},
		},
	})
}

func TestAccClusterCCSize_basic(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	cluster2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterResource(roleName, clusterName, cluster2Name, roleName, "25cc", "1", "1s", "true", "true", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "name", clusterName+"_managed"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "size", "25cc"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "replication_factor", "1"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "introspection_interval", "1s"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "introspection_debugging", "true"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "disk", "true"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "comment", "Comment"),
				),
			},
			{
				ResourceName:            "materialize_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"introspection_debugging", "introspection_interval", "identify_by_name"},
			},
		},
	})
}

func TestAccClusterManagedNoReplication_basic(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterManagedNoReplicationResource(clusterName, "3xsmall"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "name", clusterName),
					resource.TestCheckResourceAttr("materialize_cluster.test", "size", "3xsmall"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "replication_factor", "1"),
				),
			},
			{
				ResourceName:            "materialize_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"introspection_debugging", "introspection_interval", "identify_by_name"},
			},
		},
	})
}

func TestAccClusterManagedZeroReplication_basic(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterManagedZeroReplicationResource(clusterName, "3xsmall"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "name", clusterName),
					resource.TestCheckResourceAttr("materialize_cluster.test", "size", "3xsmall"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "replication_factor", "0"),
				),
			},
			{
				ResourceName:            "materialize_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"introspection_debugging", "introspection_interval", "identify_by_name"},
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterResource(roleName, oldClusterName, cluster2Name, "mz_system", "2xsmall", "2", "1s", "true", "false", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "name", oldClusterName),
					resource.TestCheckResourceAttr("materialize_cluster.test", "ownership_role", "mz_system"),
					testAccCheckClusterExists("materialize_cluster.test_role"),
					resource.TestCheckResourceAttr("materialize_cluster.test_role", "name", cluster2Name),
					resource.TestCheckResourceAttr("materialize_cluster.test_role", "ownership_role", "mz_system"),
					testAccCheckClusterExists("materialize_cluster.test_managed_cluster"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "name", oldClusterName+"_managed"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "size", "2xsmall"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "replication_factor", "2"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "introspection_interval", "1s"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "introspection_debugging", "true"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "disk", "true"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "comment", "Comment"),
				),
			},
			{
				Config: testAccClusterResource(roleName, newClusterName, cluster2Name, roleName, "3xsmall", "1", "2s", "false", "true", "New Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "name", newClusterName),
					resource.TestCheckResourceAttr("materialize_cluster.test", "ownership_role", "mz_system"),
					testAccCheckClusterExists("materialize_cluster.test_role"),
					resource.TestCheckResourceAttr("materialize_cluster.test_role", "name", cluster2Name),
					resource.TestCheckResourceAttr("materialize_cluster.test_role", "ownership_role", roleName),
					testAccCheckClusterExists("materialize_cluster.test_managed_cluster"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "name", newClusterName+"_managed"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "size", "3xsmall"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "replication_factor", "1"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "introspection_interval", "2s"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "introspection_debugging", "false"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "disk", "true"),
					resource.TestCheckResourceAttr("materialize_cluster.test_managed_cluster", "comment", "New Comment"),
				),
			},
		},
	})
}

// Ensure updates individually
func TestAccCluster_updateName(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	oldClusterName := fmt.Sprintf("old_%s", slug)
	newClusterName := fmt.Sprintf("new_%s", slug)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterManagedResource(oldClusterName, "2xsmall", "2", "1s", "true", "false", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "name", oldClusterName),
				),
			},
			{
				Config: testAccClusterManagedResource(newClusterName, "2xsmall", "2", "1s", "true", "false", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "name", newClusterName),
				),
			},
		},
	})
}

func TestAccCluster_updateSize(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterManagedResource(clusterName, "2xsmall", "2", "1s", "true", "false", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "size", "2xsmall"),
				),
			},
			{
				Config: testAccClusterManagedResource(clusterName, "3xsmall", "2", "1s", "true", "false", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "size", "3xsmall"),
				),
			},
		},
	})
}

func TestAccCluster_updateReplicationFactor(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterManagedResource(clusterName, "2xsmall", "3", "1s", "true", "false", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "replication_factor", "3"),
				),
			},
			{
				Config: testAccClusterManagedResource(clusterName, "3xsmall", "1", "1s", "true", "false", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "replication_factor", "1"),
				),
			},
		},
	})
}

func TestAccClusterWithScheduling(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	size := "3xsmall"
	rehydrationTimeEstimate := "1 hour"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterWithSchedulingConfig(clusterName, size, rehydrationTimeEstimate, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("materialize_cluster.test_scheduling", "name", clusterName),
					resource.TestCheckResourceAttr("materialize_cluster.test_scheduling", "size", size),
					resource.TestCheckResourceAttr("materialize_cluster.test_scheduling", "scheduling.0.on_refresh.0.enabled", "true"),
					resource.TestCheckResourceAttr("materialize_cluster.test_scheduling", "scheduling.0.on_refresh.0.hydration_time_estimate", rehydrationTimeEstimate),
				),
			},
		},
	})
}

func TestAccCluster_identifyByName(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllClusterDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterResourceWithNameAsId(clusterName, "3xsmall", "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test_name_as_id"),
					resource.TestCheckResourceAttr("materialize_cluster.test_name_as_id", "name", clusterName),
					resource.TestCheckResourceAttr("materialize_cluster.test_name_as_id", "identify_by_name", "true"),
					resource.TestCheckResourceAttr("materialize_cluster.test_name_as_id", "id", "aws/us-east-1:name:"+clusterName),
					resource.TestCheckResourceAttr("materialize_cluster.test_name_as_id", "size", "3xsmall"),
					resource.TestCheckResourceAttr("materialize_cluster.test_name_as_id", "replication_factor", "1"),
				),
			},
			{
				ResourceName:      "materialize_cluster.test_name_as_id",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"identify_by_name",
					"introspection_debugging",
					"introspection_interval",
				},
			},
		},
	})
}

func TestAccCluster_disappears(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	cluster2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllClusterDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterResource(roleName, clusterName, cluster2Name, roleName, "3xsmall", "1", "1s", "true", "true", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					testAccCheckObjectDisappears(materialize.MaterializeObject{ObjectType: "CLUSTER", Name: clusterName}),
					testAccCheckClusterExists("materialize_cluster.test_managed_cluster"),
					testAccCheckObjectDisappears(materialize.MaterializeObject{ObjectType: "CLUSTER", Name: clusterName + "_managed"}),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccClusterAlterGraceful(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	size := "3xsmall"
	newSize := "2xsmall"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedClusterResource(clusterName, size, "1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("materialize_cluster.test", "name", clusterName),
					resource.TestCheckResourceAttr("materialize_cluster.test", "size", size),
					resource.TestCheckResourceAttr("materialize_cluster.test", "replication_factor", "1"),
				),
			},
			{
				Config: testAccManagedClusterResourceAlterGraceful(clusterName, newSize, "1", "COMMIT"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("materialize_cluster.test", "name", clusterName),
					resource.TestCheckResourceAttr("materialize_cluster.test", "size", newSize),
					resource.TestCheckResourceAttr("materialize_cluster.test", "replication_factor", "1"),
				),
			},
		},
	})
}

func testAccClusterResource(
	roleName,
	cluster1Name,
	cluster2Name,
	cluster2Owner,
	clusterSize,
	clusterReplicationFactor,
	introspectionInterval,
	introspectionDebugging,
	disk,
	comment string,
) string {
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
		name                          = "%[2]s_managed"
		size                          = "%[5]s"
		replication_factor            = %[6]s
		introspection_interval        = "%[7]s"
		introspection_debugging       = %[8]s
		disk                          = %[9]s
		comment                       = "%[10]s"
		availability_zones            = ["test1", "test2"]
	}
	`,
		roleName,
		cluster1Name,
		cluster2Name,
		cluster2Owner,
		clusterSize,
		clusterReplicationFactor,
		introspectionInterval,
		introspectionDebugging,
		disk,
		comment)
}

func testAccClusterManagedNoReplicationResource(clusterName, clusterSize string) string {
	return fmt.Sprintf(`
	resource "materialize_cluster" "test" {
		name = "%[1]s"
		size = "%[2]s"
	}
	`,
		clusterName, clusterSize)
}

func testAccManagedClusterResource(clusterName, clusterSize string, replicationFactor string) string {
	return fmt.Sprintf(`
	resource "materialize_cluster" "test" {
		name = "%[1]s"
		size = "%[2]s"
		replication_factor = %[3]s
	}
	`,
		clusterName, clusterSize, replicationFactor)
}

func testAccManagedClusterResourceAlterGraceful(clusterName, clusterSize string, replicationFactor string, onTimeoutAction string) string {
	return fmt.Sprintf(`
	resource "materialize_cluster" "test" {
		name = "%[1]s"
		size = "%[2]s"
		replication_factor  = %[3]s
		wait_until_ready {
			enabled = true
			timeout = "10m"
			on_timeout = "%[4]s"
		}
	}
	`,
		clusterName, clusterSize, replicationFactor, onTimeoutAction)
}

func testAccClusterManagedZeroReplicationResource(clusterName, clusterSize string) string {
	return fmt.Sprintf(`
	resource "materialize_cluster" "test" {
		name = "%[1]s"
		size = "%[2]s"
		replication_factor = 0
	}
	`,
		clusterName, clusterSize)
}

func testAccClusterManagedResource(
	clusterName,
	clusterSize,
	clusterReplicationFactor,
	introspectionInterval,
	introspectionDebugging,
	disk,
	comment string) string {
	return fmt.Sprintf(`
	resource "materialize_cluster" "test" {
		name                          = "%[1]s"
		size                          = "%[2]s"
		replication_factor            = %[3]s
		introspection_interval        = "%[4]s"
		introspection_debugging       = %[5]s
		disk                          = %[6]s
		comment                       = "%[7]s"
	}
	`,
		clusterName,
		clusterSize,
		clusterReplicationFactor,
		introspectionInterval,
		introspectionDebugging,
		disk,
		comment)
}

func testAccClusterWithSchedulingConfig(clusterName, size, rehydrationTimeEstimate string, onRefresh bool) string {
	onRefreshStr := "false"
	if onRefresh {
		onRefreshStr = "true"
	}
	return fmt.Sprintf(`
resource "materialize_cluster" "test_scheduling" {
    name                    = "%s"
    size                    = "%s"
    scheduling {
        on_refresh {
			enabled = %s
			hydration_time_estimate = "%s"
		}
    }
}
`, clusterName, size, onRefreshStr, rehydrationTimeEstimate)
}

func testAccClusterResourceWithNameAsId(clusterName, clusterSize, clusterReplicationFactor string) string {
	return fmt.Sprintf(`
	resource "materialize_cluster" "test_name_as_id" {
		name                = "%[1]s"
		size                = "%[2]s"
		replication_factor  = %[3]s
		identify_by_name      = true
	}
	`,
		clusterName,
		clusterSize,
		clusterReplicationFactor)
}

func testAccCheckClusterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("cluster not found: %s", name)
		}
		identifyByName := false
		if r.Primary.Attributes["identify_by_name"] == "true" {
			identifyByName = true
		}
		_, err = materialize.ScanCluster(db, utils.ExtractId(r.Primary.ID), identifyByName)
		return err
	}
}

func testAccCheckAllClusterDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_cluster" {
			continue
		}

		_, err := materialize.ScanCluster(db, utils.ExtractId(r.Primary.ID), false)
		if err == nil {
			return fmt.Errorf("Cluster %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

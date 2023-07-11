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

func TestAccGrantCluster_basic(t *testing.T) {
	privilege := randomPrivilege("CLUSTER")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantClusterResource(roleName, clusterName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantClusterExists("materialize_cluster_grant.cluster_grant", roleName, clusterName, privilege),
					resource.TestCheckResourceAttr("materialize_cluster_grant.cluster_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_cluster_grant.cluster_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_cluster_grant.cluster_grant", "cluster_name", clusterName),
				),
			},
		},
	})
}

func TestAccGrantCluster_disappears(t *testing.T) {
	privilege := randomPrivilege("CLUSTER")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantClusterResource(roleName, clusterName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantClusterExists("materialize_cluster_grant.cluster_grant", roleName, clusterName, privilege),
					testAccCheckGrantClusterRevoked(roleName, clusterName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantClusterResource(roleName, clusterName, privilege string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%s"
}

resource "materialize_cluster" "cluster" {
	name = "%s"
}

resource "materialize_cluster_grant" "cluster_grant" {
	role_name    = materialize_role.test.name
	privilege    = "%s"
	cluster_name = materialize_cluster.cluster.name
}
`, roleName, clusterName, privilege)
}

func testAccCheckGrantClusterExists(grantName, roleName, clusterName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("grant not found")
		}

		o := materialize.ObjectSchemaStruct{Name: clusterName}
		id, err := materialize.ClusterId(db, o)
		if err != nil {
			return err
		}

		roleId, err := materialize.RoleId(db, roleName)
		if err != nil {
			return err
		}

		g, err := materialize.ScanPrivileges(db, "CLUSTER", id)
		if err != nil {
			return err
		}

		privilegeMap := materialize.ParsePrivileges(g)
		if !materialize.HasPrivilege(privilegeMap[roleId], privilege) {
			return fmt.Errorf("cluster object %s does not include privilege %s", g, privilege)
		}
		return nil
	}
}

func testAccCheckGrantClusterRevoked(roleName, clusterName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`REVOKE %s ON CLUSTER "%s" FROM "%s";`, privilege, clusterName, roleName))
		return err
	}
}

package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					testAccCheckGrantExists(
						materialize.MaterializeObject{
							ObjectType: "CLUSTER",
							Name:       clusterName,
						}, "materialize_cluster_grant.cluster_grant", roleName, privilege),
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

	o := materialize.MaterializeObject{
		ObjectType: "CLUSTER",
		Name:       clusterName,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantClusterResource(roleName, clusterName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(o, "materialize_cluster_grant.cluster_grant", roleName, privilege),
					testAccCheckGrantRevoked(o, roleName, privilege),
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

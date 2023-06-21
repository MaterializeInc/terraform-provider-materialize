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

func TestAccCluster_basic(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterResource(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					resource.TestCheckResourceAttr("materialize_cluster.test", "name", clusterName),
					resource.TestCheckResourceAttr("materialize_cluster.test", "ownership_role", "mz_system"),
				),
			},
		},
	})
}

func TestAccCluster_disappears(t *testing.T) {
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterResource(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("materialize_cluster.test"),
					testAccCheckClusterDisappears(clusterName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccClusterResource(clusterName string) string {
	return fmt.Sprintf(`
resource "materialize_cluster" "test" {
	name = "%[1]s"
}
`, clusterName)
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

func testAccCheckClusterDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP CLUSTER "%s";`, name))
		return err
	}
}

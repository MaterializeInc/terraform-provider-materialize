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

func TestAccMaterializedView_basic(t *testing.T) {
	viewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccMaterializedViewResource(viewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaterializedViewExists("materialize_materialized_view.test"),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "name", viewName),
				),
			},
		},
	})
}

func TestAccMaterializedView_update(t *testing.T) {
	viewName := "old_mz_view"
	newViewName := "new_mz_view"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccMaterializedViewResource(viewName),
			},
			{
				Config: testAccMaterializedViewResource(newViewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaterializedViewExists("materialize_materialized_view.test"),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "name", newViewName),
				),
			},
		},
	})
}

func TestAccMaterializedView_disappears(t *testing.T) {
	viewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllMaterializedViewsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccMaterializedViewResource(viewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaterializedViewExists("materialize_materialized_view.test"),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "name", viewName),
					testAccCheckMaterializedViewDisappears(viewName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccMaterializedViewResource(name string) string {
	return fmt.Sprintf(`
resource "materialize_materialized_view" "test" {
	name = "%s"
	statement = "SELECT 1 AS id"
}
`, name)
}

func testAccCheckMaterializedViewExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Materialized View not found: %s", name)
		}
		_, err := materialize.ScanMaterializedView(db, r.Primary.ID)
		return err
	}
}

func testAccCheckMaterializedViewDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP MATERIALIZED VIEW "%s";`, name))
		return err
	}
}

func testAccCheckAllMaterializedViewsDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_materialized_view" {
			continue
		}

		_, err := materialize.ScanMaterializedView(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("Materialized View %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}
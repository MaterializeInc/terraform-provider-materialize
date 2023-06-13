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

func TestAccView_basic(t *testing.T) {
	viewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccViewResource(viewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckViewExists("materialize_view.test"),
					resource.TestCheckResourceAttr("materialize_view.test", "name", viewName),
				),
			},
		},
	})
}

func TestAccView_update(t *testing.T) {
	viewName := "old_view"
	newViewName := "new_view"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccViewResource(viewName),
			},
			{
				Config: testAccViewResource(newViewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckViewExists("materialize_view.test"),
					resource.TestCheckResourceAttr("materialize_view.test", "name", newViewName),
				),
			},
		},
	})
}

func TestAccView_disappears(t *testing.T) {
	viewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllViewsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccViewResource(viewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckViewExists("materialize_view.test"),
					resource.TestCheckResourceAttr("materialize_view.test", "name", viewName),
					testAccCheckViewDisappears(viewName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccViewResource(name string) string {
	return fmt.Sprintf(`
resource "materialize_view" "test" {
	name = "%s"
	statement = "SELECT 1 AS id"
}
`, name)
}

func testAccCheckViewExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("View not found: %s", name)
		}
		_, err := materialize.ScanView(db, r.Primary.ID)
		return err
	}
}

func testAccCheckViewDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP VIEW "%s";`, name))
		return err
	}
}

func testAccCheckAllViewsDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_view" {
			continue
		}

		_, err := materialize.ScanView(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("View %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

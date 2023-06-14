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

func TestAccDatabase_basic(t *testing.T) {
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseResource(databaseName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("materialize_database.test"),
					resource.TestCheckResourceAttr("materialize_database.test", "name", databaseName),
				),
			},
		},
	})
}

func TestAccDatabase_disappears(t *testing.T) {
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllDatabasesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseResource(databaseName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("materialize_database.test"),
					testAccCheckDatabaseDisappears(databaseName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDatabaseResource(name string) string {
	return fmt.Sprintf(`
resource "materialize_database" "test" {
	name = "%s"
}
`, name)
}

func testAccCheckDatabaseExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("database not found: %s", name)
		}
		_, err := materialize.ScanDatabase(db, r.Primary.ID)
		return err
	}
}

func testAccCheckDatabaseDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP DATABASE "%s";`, name))
		return err
	}
}

func testAccCheckAllDatabasesDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_database" {
			continue
		}

		_, err := materialize.ScanDatabase(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("database %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

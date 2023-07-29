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
	for _, roleName := range []string{
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha),
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + "@materialize.com",
	} {
		t.Run(fmt.Sprintf("roleName=%s", roleName), func(t *testing.T) {
			databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
			database2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
			resource.ParallelTest(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: testAccProviderFactories,
				CheckDestroy:      nil,
				Steps: []resource.TestStep{
					{
						Config: testAccDatabaseResource(roleName, databaseName, database2Name, roleName),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckDatabaseExists("materialize_database.test"),
							resource.TestCheckResourceAttr("materialize_database.test", "name", databaseName),
							resource.TestCheckResourceAttr("materialize_database.test", "ownership_role", "mz_system"),
							testAccCheckDatabaseExists("materialize_database.test_role"),
							resource.TestCheckResourceAttr("materialize_database.test_role", "name", database2Name),
							resource.TestCheckResourceAttr("materialize_database.test_role", "ownership_role", roleName),
						),
					},
					{
						ResourceName:      "materialize_database.test",
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

func TestAccDatabase_disappears(t *testing.T) {
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	database2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllDatabasesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseResource(roleName, databaseName, database2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("materialize_database.test"),
					testAccCheckObjectDisappears(
						materialize.ObjectSchemaStruct{
							ObjectType: "DATABASE",
							Name:       databaseName,
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDatabaseResource(roleName, databaseName, databse2Name, databaseOwner string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_database" "test" {
	name = "%[2]s"
}

resource "materialize_database" "test_role" {
	name = "%[3]s"
	ownership_role = "%[4]s"

	depends_on = [materialize_role.test]
}
`, roleName, databaseName, databse2Name, databaseOwner)
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

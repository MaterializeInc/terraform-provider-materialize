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

func TestAccTable_basic(t *testing.T) {
	tableName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	tableRoleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTableResource(roleName, tableName, tableRoleName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists("materialize_table.test"),
					resource.TestCheckResourceAttr("materialize_table.test", "name", tableName),
					resource.TestCheckResourceAttr("materialize_table.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_table.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_table.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_table.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, tableName)),
					resource.TestCheckResourceAttr("materialize_table.test", "column.#", "3"),
					resource.TestCheckResourceAttr("materialize_table.test", "ownership_role", "mz_system"),
					testAccCheckTableExists("materialize_table.test_role"),
					resource.TestCheckResourceAttr("materialize_table.test_role", "name", tableRoleName),
					resource.TestCheckResourceAttr("materialize_table.test_role", "ownership_role", roleName),
				),
			},
			{
				ResourceName:      "materialize_table.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTable_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	tableName := fmt.Sprintf("old_%s", slug)
	newTableName := fmt.Sprintf("new_%s", slug)
	tableRoleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTableResource(roleName, tableName, tableRoleName, "mz_system"),
			},
			{
				Config: testAccTableResource(roleName, newTableName, tableRoleName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists("materialize_table.test"),
					resource.TestCheckResourceAttr("materialize_table.test", "name", newTableName),
					resource.TestCheckResourceAttr("materialize_table.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_table.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_table.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_table.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newTableName)),
					resource.TestCheckResourceAttr("materialize_table.test", "column.#", "3"),
					testAccCheckTableExists("materialize_table.test_role"),
					resource.TestCheckResourceAttr("materialize_table.test_role", "name", tableRoleName),
					resource.TestCheckResourceAttr("materialize_table.test_role", "ownership_role", roleName),
				),
			},
		},
	})
}

func TestAccTable_disappears(t *testing.T) {
	tableName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	tableRoleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllTablesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccTableResource(roleName, tableName, tableRoleName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists("materialize_table.test"),
					resource.TestCheckResourceAttr("materialize_table.test", "name", tableName),
					resource.TestCheckResourceAttr("materialize_table.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_table.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_table.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, tableName)),
					resource.TestCheckResourceAttr("materialize_table.test", "column.#", "3"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "TABLE",
							Name:       tableName,
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTableResource(roleName, tableName, tableRoleName, tableOwnership string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%s"
}

resource "materialize_table" "test" {
	name = "%s"
	column {
		name = "column_1"
		type = "text"
	}
	column {
		name = "column_2"
		type = "int"
	}
	column {
		name     = "column_3"
		type     = "text"
		nullable = true
	}
}

resource "materialize_table" "test_role" {
	name = "%s"
	ownership_role = "%s"

	column {
		name = "column_1"
		type = "text"
	}

	depends_on = [materialize_role.test]
}
`, roleName, tableName, tableRoleName, tableOwnership)
}

func testAccCheckTableExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Table not found: %s", name)
		}
		_, err := materialize.ScanTable(db, r.Primary.ID)
		return err
	}
}

func testAccCheckAllTablesDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_table" {
			continue
		}

		_, err := materialize.ScanTable(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("Table %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

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
				Config: testAccTableResource(roleName, tableName, tableRoleName, roleName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists("materialize_table.test"),
					resource.TestMatchResourceAttr("materialize_table.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_table.test", "name", tableName),
					resource.TestCheckResourceAttr("materialize_table.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_table.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_table.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_table.test", "comment", "comment"),
					resource.TestCheckResourceAttr("materialize_table.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, tableName)),
					resource.TestCheckResourceAttr("materialize_table.test", "column.#", "5"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.0.name", "column_1"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.0.type", "text"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.0.nullable", "false"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.0.default", "NULL"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.1.name", "column_2"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.1.type", "integer"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.2.name", "column_3"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.2.nullable", "true"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.3.name", "column_4"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.3.default", "NULL"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.4.name", "column_5"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.4.default", "NULL"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.4.default", "NULL"),
					resource.TestCheckResourceAttr("materialize_table.test", "ownership_role", "mz_system"),
					testAccCheckTableExists("materialize_table.test_role"),
					resource.TestCheckResourceAttr("materialize_table.test_role", "name", tableRoleName),
					resource.TestCheckResourceAttr("materialize_table.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_table.test_role", "comment", "Comment"),
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
				Config: testAccTableResource(roleName, tableName, tableRoleName, "mz_system", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists("materialize_table.test"),
					testAccCheckTableExists("materialize_table.test_role"),
					resource.TestCheckResourceAttr("materialize_table.test_role", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_table.test_role", "comment", "Comment"),
				),
			},
			{
				Config: testAccTableResource(roleName, newTableName, tableRoleName, roleName, "New Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists("materialize_table.test"),
					resource.TestCheckResourceAttr("materialize_table.test", "name", newTableName),
					resource.TestCheckResourceAttr("materialize_table.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_table.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_table.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_table.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newTableName)),
					resource.TestCheckResourceAttr("materialize_table.test", "column.#", "5"),
					testAccCheckTableExists("materialize_table.test_role"),
					resource.TestCheckResourceAttr("materialize_table.test_role", "name", tableRoleName),
					resource.TestCheckResourceAttr("materialize_table.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_table.test_role", "comment", "New Comment"),
				),
			},
			{
				Config: testAccTableResourceWithUpdates(roleName, tableName, tableRoleName, "mz_system", "new_column_1", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists("materialize_table.test"),
					resource.TestCheckResourceAttr("materialize_table.test", "name", tableName),
					resource.TestCheckResourceAttr("materialize_table.test", "column.0.name", "new_column_1"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.0.type", "text"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.1.name", "column_2"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.1.type", "int"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.2.name", "column_3"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.2.type", "text"),
					resource.TestCheckResourceAttr("materialize_table.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_table.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_table.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_table.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, tableName)),
					resource.TestCheckResourceAttr("materialize_table.test", "column.#", "5"),
					testAccCheckTableExists("materialize_table.test_role"),
					resource.TestCheckResourceAttr("materialize_table.test_role", "name", tableRoleName),
					resource.TestCheckResourceAttr("materialize_table.test_role", "ownership_role", roleName),
				),
				ResourceName:      "materialize_table.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableResourceWithUpdates(roleName, tableName, tableRoleName, "mz_system", "", "Updated comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists("materialize_table.test"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.1.comment", "Updated comment"),
					resource.TestCheckResourceAttr("materialize_table.test", "name", tableName),
					resource.TestCheckResourceAttr("materialize_table.test", "column.0.name", "column_1"),
					resource.TestCheckResourceAttr("materialize_table.test", "column.0.type", "text"),
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
				Config: testAccTableResource(roleName, tableName, tableRoleName, roleName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists("materialize_table.test"),
					resource.TestCheckResourceAttr("materialize_table.test", "name", tableName),
					resource.TestCheckResourceAttr("materialize_table.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_table.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_table.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, tableName)),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "TABLE",
							Name:       tableName,
						},
					),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTableResource(roleName, tableName, tableRoleName, tableOwnership, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "test" {
		name = "%[1]s"
	}

	resource "materialize_table" "test" {
		name = "%[2]s"
		comment = "comment"
		column {
			name = "column_1"
			type = "text"
		}
		column {
			name    = "column_2"
			type    = "int"
			comment = "comment"
		}
		column {
			name     = "column_3"
			type     = "text"
			nullable = true
		}
		column {
			name    = "column_4"
			type    = "text"
			default = "NULL"
		}
		column {
			name     = "column_5"
			type     = "text"
			nullable = true
			default  = "NULL"
		}
	}

	resource "materialize_table" "test_role" {
		name = "%[3]s"
		ownership_role = "%[4]s"
		comment = "%[5]s"

		column {
			name = "column_1"
			type = "text"
		}

		depends_on = [materialize_role.test]
	}
	`, roleName, tableName, tableRoleName, tableOwnership, comment)
}

func testAccCheckTableExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Table not found: %s", name)
		}
		_, err = materialize.ScanTable(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllTablesDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_table" {
			continue
		}

		_, err := materialize.ScanTable(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("Table %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

func testAccTableResourceWithUpdates(roleName, tableName, tableRoleName, tableOwnership, newColumnName, updatedComment string) string {
	columnName1 := "column_1"
	if newColumnName != "" {
		columnName1 = newColumnName
	}

	commentColumn2 := "comment"
	if updatedComment != "" {
		commentColumn2 = updatedComment
	}

	return fmt.Sprintf(`
	resource "materialize_role" "test" {
		name = "%[1]s"
	}

	resource "materialize_table" "test" {
		name = "%[2]s"
		comment = "Initial table comment"
		column {
			name = "%[3]s"
			type = "text"
		}
		column {
			name    = "column_2"
			type    = "int"
			comment = "%[4]s"
		}
		column {
			name     = "column_3"
			type     = "text"
			nullable = true
		}
		ownership_role = "%[5]s"
	}

	resource "materialize_table" "test_role" {
		name = "%[6]s"
		ownership_role = "%[7]s"

		column {
			name = "%[3]s"
			type = "text"
		}

		depends_on = [materialize_role.test]
	}
	`, roleName, tableName, columnName1, commentColumn2, tableOwnership, tableRoleName, tableOwnership)
}

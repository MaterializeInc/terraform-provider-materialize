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

func TestAccGrantTable_basic(t *testing.T) {
	privilege := randomPrivilege("TABLE")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	tableName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantTableResource(roleName, tableName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantTableExists("materialize_grant_table.table_grant", roleName, tableName, schemaName, databaseName, privilege),
					resource.TestCheckResourceAttr("materialize_grant_table.table_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_grant_table.table_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_grant_table.table_grant", "table_name", tableName),
					resource.TestCheckResourceAttr("materialize_grant_table.table_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_grant_table.table_grant", "database_name", databaseName),
				),
			},
		},
	})
}

func TestAccGrantTable_disappears(t *testing.T) {
	privilege := randomPrivilege("TABLE")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	tableName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantTableResource(roleName, tableName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantTableExists("materialize_grant_table.table_grant", roleName, tableName, schemaName, databaseName, privilege),
					testAccCheckGrantTableRevoked(roleName, tableName, schemaName, databaseName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantTableResource(roleName, tableName, schemaName, databaseName, privilege string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%s"
}

resource "materialize_database" "test" {
	name = "%s"
}

resource "materialize_schema" "test" {
	name = "%s"
	database_name = materialize_database.test.name
}

resource "materialize_table" "test" {
	name          = "%s"
	schema_name   = materialize_schema.test.name
	database_name = materialize_database.test.name
  
	column {
	  name = "column_1"
	  type = "text"
	}
}

resource "materialize_grant_table" "table_grant" {
	role_name     = materialize_role.test.name
	privilege     = "%s"
	database_name = materialize_database.test.name
	schema_name   = materialize_schema.test.name
	table_name    = materialize_table.test.name
}
`, roleName, databaseName, schemaName, tableName, privilege)
}

func testAccCheckGrantTableExists(grantName, roleName, tableName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("grant not found")
		}

		o := materialize.ObjectSchemaStruct{Name: tableName, SchemaName: schemaName, DatabaseName: databaseName}
		id, err := materialize.TableId(db, o)
		if err != nil {
			return err
		}

		roleId, err := materialize.RoleId(db, roleName)
		if err != nil {
			return err
		}

		g, err := materialize.ScanPrivileges(db, "TABLE", id)
		if err != nil {
			return err
		}

		privilegeMap := materialize.ParsePrivileges(g)
		if !materialize.HasPrivilege(privilegeMap[roleId], privilege) {
			return fmt.Errorf("schema object %s does not include privilege %s", g, privilege)
		}
		return nil
	}
}

func testAccCheckGrantTableRevoked(roleName, tableName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`REVOKE %s ON TABLE "%s"."%s"."%s" FROM "%s";`, privilege, databaseName, schemaName, tableName, roleName))
		return err
	}
}

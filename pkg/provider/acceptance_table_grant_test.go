package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					testAccCheckGrantExists(
						materialize.ObjectSchemaStruct{
							ObjectType:   "TABLE",
							Name:         tableName,
							SchemaName:   schemaName,
							DatabaseName: databaseName,
						}, "materialize_table_grant.table_grant", roleName, privilege),
					resource.TestCheckResourceAttr("materialize_table_grant.table_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_table_grant.table_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_table_grant.table_grant", "table_name", tableName),
					resource.TestCheckResourceAttr("materialize_table_grant.table_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_table_grant.table_grant", "database_name", databaseName),
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

	o := materialize.ObjectSchemaStruct{
		ObjectType:   "TABLE",
		Name:         tableName,
		SchemaName:   schemaName,
		DatabaseName: databaseName,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantTableResource(roleName, tableName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(o, "materialize_table_grant.table_grant", roleName, privilege),
					testAccCheckGrantRevoked(o, roleName, privilege),
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

resource "materialize_table_grant" "table_grant" {
	role_name     = materialize_role.test.name
	privilege     = "%s"
	database_name = materialize_database.test.name
	schema_name   = materialize_schema.test.name
	table_name    = materialize_table.test.name
}
`, roleName, databaseName, schemaName, tableName, privilege)
}

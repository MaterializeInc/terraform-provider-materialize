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

func TestAccGrantType_basic(t *testing.T) {
	privilege := randomPrivilege("TYPE")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	typeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantTypeResource(roleName, typeName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantTypeExists("materialize_grant_type.type_grant", roleName, typeName, schemaName, databaseName, privilege),
					resource.TestCheckResourceAttr("materialize_grant_type.type_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_grant_type.type_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_grant_type.type_grant", "type_name", typeName),
					resource.TestCheckResourceAttr("materialize_grant_type.type_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_grant_type.type_grant", "database_name", databaseName),
				),
			},
		},
	})
}

func TestAccGrantType_disappears(t *testing.T) {
	privilege := randomPrivilege("TYPE")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	typeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantTypeResource(roleName, typeName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantTypeExists("materialize_grant_type.type_grant", roleName, typeName, schemaName, databaseName, privilege),
					testAccCheckGrantTypeRevoked(roleName, typeName, schemaName, databaseName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantTypeResource(roleName, typeName, schemaName, databaseName, privilege string) string {
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

resource "materialize_type" "test" {
	name          = "%s"
	schema_name   = materialize_schema.test.name
	database_name = materialize_database.test.name
  
	list_properties {
	  element_type = "int4"
	}
}

resource "materialize_grant_type" "type_grant" {
	role_name     = materialize_role.test.name
	privilege     = "%s"
	database_name = materialize_database.test.name
	schema_name   = materialize_schema.test.name
	type_name     = materialize_type.test.name
}
`, roleName, databaseName, schemaName, typeName, privilege)
}

func testAccCheckGrantTypeExists(grantName, roleName, typeName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("grant not found")
		}

		o := materialize.ObjectSchemaStruct{Name: typeName, SchemaName: schemaName, DatabaseName: databaseName}
		id, err := materialize.TypeId(db, o)
		if err != nil {
			return err
		}

		roleId, err := materialize.RoleId(db, roleName)
		if err != nil {
			return err
		}

		g, err := materialize.ScanPrivileges(db, "TYPE", id)
		if err != nil {
			return err
		}

		privilegeMap := materialize.ParsePrivileges(g)
		if !materialize.HasPrivilege(privilegeMap[roleId], privilege) {
			return fmt.Errorf("type object %s does not include privilege %s", g, privilege)
		}
		return nil
	}
}

func testAccCheckGrantTypeRevoked(roleName, typeName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`REVOKE %s ON TYPE "%s"."%s"."%s" FROM "%s";`, privilege, databaseName, schemaName, typeName, roleName))
		return err
	}
}

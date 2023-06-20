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

func TestAccGrantSchema_basic(t *testing.T) {
	privilege := randomPrivilege("SCHEMA")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantSchemaResource(roleName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantSchemaExists("materialize_grant_schema.schema_grant", roleName, schemaName, databaseName, privilege),
					resource.TestCheckResourceAttr("materialize_grant_schema.schema_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_grant_schema.schema_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_grant_schema.schema_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_grant_schema.schema_grant", "database_name", databaseName),
				),
			},
		},
	})
}

func TestAccGrantSchema_disappears(t *testing.T) {
	privilege := randomPrivilege("SCHEMA")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantSchemaResource(roleName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantSchemaExists("materialize_grant_schema.schema_grant", roleName, schemaName, databaseName, privilege),
					testAccCheckGrantSchemaRevoked(roleName, schemaName, databaseName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantSchemaResource(roleName, schemaName, databaseName, privilege string) string {
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

resource "materialize_grant_schema" "schema_grant" {
	role_name     = materialize_role.test.name
	privilege     = "%s"
	database_name = materialize_database.test.name
	schema_name   = materialize_schema.test.name
}
`, roleName, databaseName, schemaName, privilege)
}

func testAccCheckGrantSchemaExists(grantName, roleName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("grant not found")
		}

		schemaId, err := materialize.SchemaId(db, schemaName, databaseName)
		if err != nil {
			return err
		}

		roleId, err := materialize.RoleId(db, roleName)
		if err != nil {
			return err
		}

		g, err := materialize.ScanPrivileges(db, "SCHEMA", schemaId)
		if err != nil {
			return err
		}

		priviledgeMap := materialize.ParsePrivileges(g)
		if !materialize.HasPrivilege(priviledgeMap[roleId], privilege) {
			return fmt.Errorf("schema object %s does not include privilege %s", g, privilege)
		}
		return nil
	}
}

func testAccCheckGrantSchemaRevoked(roleName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`REVOKE %s ON SCHEMA "%s"."%s" FROM "%s";`, privilege, databaseName, schemaName, roleName))
		return err
	}
}

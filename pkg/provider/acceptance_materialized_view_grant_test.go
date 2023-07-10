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

func TestAccGrantMaterializedView_basic(t *testing.T) {
	privilege := randomPrivilege("MATERIALIZED VIEW")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	materializedViewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantMaterializedViewResource(roleName, materializedViewName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantMaterializedViewExists("materialize_materialized_view_grant.materialized_view_grant", roleName, materializedViewName, schemaName, databaseName, privilege),
					resource.TestCheckResourceAttr("materialize_materialized_view_grant.materialized_view_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_materialized_view_grant.materialized_view_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_materialized_view_grant.materialized_view_grant", "materialized_view_name", materializedViewName),
					resource.TestCheckResourceAttr("materialize_materialized_view_grant.materialized_view_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_materialized_view_grant.materialized_view_grant", "database_name", databaseName),
				),
			},
		},
	})
}

func TestAccGrantMaterializedView_disappears(t *testing.T) {
	privilege := randomPrivilege("MATERIALIZED VIEW")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	materializedViewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantMaterializedViewResource(roleName, materializedViewName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantMaterializedViewExists("materialize_materialized_view_grant.materialized_view_grant", roleName, materializedViewName, schemaName, databaseName, privilege),
					testAccCheckGrantMaterializedViewRevoked(roleName, materializedViewName, schemaName, databaseName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantMaterializedViewResource(roleName, materializedViewName, schemaName, databaseName, privilege string) string {
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

resource "materialize_materialized_view" "test" {
	name          = "%s"
	schema_name   = materialize_schema.test.name
	database_name = materialize_database.test.name
	cluster_name  = "default"
  
	statement = <<SQL
  SELECT
	  1 AS id
  SQL
}

resource "materialize_materialized_view_grant" "materialized_view_grant" {
	role_name              = materialize_role.test.name
	privilege              = "%s"
	database_name          = materialize_database.test.name
	schema_name            = materialize_schema.test.name
	materialized_view_name = materialize_materialized_view.test.name
}
`, roleName, databaseName, schemaName, materializedViewName, privilege)
}

func testAccCheckGrantMaterializedViewExists(grantName, roleName, materializedViewName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("grant not found")
		}

		o := materialize.ObjectSchemaStruct{Name: materializedViewName, SchemaName: schemaName, DatabaseName: databaseName}
		id, err := materialize.MaterializedViewId(db, o)
		if err != nil {
			return err
		}

		roleId, err := materialize.RoleId(db, roleName)
		if err != nil {
			return err
		}

		g, err := materialize.ScanPrivileges(db, "MATERIALIZED VIEW", id)
		if err != nil {
			return err
		}

		privilegeMap := materialize.ParsePrivileges(g)
		if !materialize.HasPrivilege(privilegeMap[roleId], privilege) {
			return fmt.Errorf("materialized view object %s does not include privilege %s", g, privilege)
		}
		return nil
	}
}

func testAccCheckGrantMaterializedViewRevoked(roleName, materializedViewName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`REVOKE %s ON MATERIALIZED VIEW "%s"."%s"."%s" FROM "%s";`, privilege, databaseName, schemaName, materializedViewName, roleName))
		return err
	}
}

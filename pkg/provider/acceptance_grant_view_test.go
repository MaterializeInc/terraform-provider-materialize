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

func TestAccGrantView_basic(t *testing.T) {
	privilege := randomPrivilege("VIEW")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	viewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantViewResource(roleName, viewName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantViewExists("materialize_grant_view.view_grant", roleName, viewName, schemaName, databaseName, privilege),
					resource.TestCheckResourceAttr("materialize_grant_view.view_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_grant_view.view_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_grant_view.view_grant", "view_name", viewName),
					resource.TestCheckResourceAttr("materialize_grant_view.view_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_grant_view.view_grant", "database_name", databaseName),
				),
			},
		},
	})
}

func TestAccGrantView_disappears(t *testing.T) {
	privilege := randomPrivilege("VIEW")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	viewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantViewResource(roleName, viewName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantViewExists("materialize_grant_view.view_grant", roleName, viewName, schemaName, databaseName, privilege),
					testAccCheckGrantViewRevoked(roleName, viewName, schemaName, databaseName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantViewResource(roleName, viewName, schemaName, databaseName, privilege string) string {
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

resource "materialize_view" "test" {
	name          = "%s"
	schema_name   = materialize_schema.test.name
	database_name = materialize_database.test.name
  
	statement = <<SQL
  SELECT
	  1 AS id
  SQL
}

resource "materialize_grant_view" "view_grant" {
	role_name     = materialize_role.test.name
	privilege     = "%s"
	database_name = materialize_database.test.name
	schema_name   = materialize_schema.test.name
	view_name     = materialize_view.test.name
}
`, roleName, databaseName, schemaName, viewName, privilege)
}

func testAccCheckGrantViewExists(grantName, roleName, viewName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("grant not found")
		}

		id, err := materialize.ViewId(db, viewName, schemaName, databaseName)
		if err != nil {
			return err
		}

		roleId, err := materialize.RoleId(db, roleName)
		if err != nil {
			return err
		}

		g, err := materialize.ScanPrivileges(db, "VIEW", id)
		if err != nil {
			return err
		}

		privilegeMap := materialize.ParsePrivileges(g)
		if !materialize.HasPrivilege(privilegeMap[roleId], privilege) {
			return fmt.Errorf("view object %s does not include privilege %s", g, privilege)
		}
		return nil
	}
}

func testAccCheckGrantViewRevoked(roleName, viewName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`REVOKE %s ON VIEW "%s"."%s"."%s" FROM "%s";`, privilege, databaseName, schemaName, viewName, roleName))
		return err
	}
}

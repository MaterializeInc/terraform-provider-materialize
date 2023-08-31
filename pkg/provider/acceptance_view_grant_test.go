package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					testAccCheckGrantExists(
						materialize.MaterializeObject{
							ObjectType:   "VIEW",
							Name:         viewName,
							SchemaName:   schemaName,
							DatabaseName: databaseName,
						}, "materialize_view_grant.view_grant", roleName, privilege),
					resource.TestCheckResourceAttr("materialize_view_grant.view_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_view_grant.view_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_view_grant.view_grant", "view_name", viewName),
					resource.TestCheckResourceAttr("materialize_view_grant.view_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_view_grant.view_grant", "database_name", databaseName),
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

	o := materialize.MaterializeObject{
		ObjectType:   "VIEW",
		Name:         viewName,
		SchemaName:   schemaName,
		DatabaseName: databaseName,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantViewResource(roleName, viewName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(o, "materialize_view_grant.view_grant", roleName, privilege),
					testAccCheckGrantRevoked(o, roleName, privilege),
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

resource "materialize_view_grant" "view_grant" {
	role_name     = materialize_role.test.name
	privilege     = "%s"
	database_name = materialize_database.test.name
	schema_name   = materialize_schema.test.name
	view_name     = materialize_view.test.name
}
`, roleName, databaseName, schemaName, viewName, privilege)
}

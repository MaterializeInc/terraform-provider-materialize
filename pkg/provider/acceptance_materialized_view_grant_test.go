package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					testAccCheckGrantExists(
						materialize.MaterializeObject{
							ObjectType:   "MATERIALIZED VIEW",
							Name:         materializedViewName,
							SchemaName:   schemaName,
							DatabaseName: databaseName,
						}, "materialize_materialized_view_grant.materialized_view_grant", roleName, privilege),
					resource.TestMatchResourceAttr("materialize_materialized_view_grant.materialized_view_grant", "id", terraformGrantIdRegex),
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

	o := materialize.MaterializeObject{
		ObjectType:   "MATERIALIZED VIEW",
		Name:         materializedViewName,
		SchemaName:   schemaName,
		DatabaseName: databaseName,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantMaterializedViewResource(roleName, materializedViewName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(o, "materialize_materialized_view_grant.materialized_view_grant", roleName, privilege),
					testAccCheckGrantRevoked(o, roleName, privilege),
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
	name = "%s"
	schema_name = materialize_schema.test.name
	database_name = materialize_database.test.name
	cluster_name = "default"
  
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

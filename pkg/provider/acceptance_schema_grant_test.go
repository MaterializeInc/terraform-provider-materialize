package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					testAccCheckGrantExists(materialize.MaterializeObject{
						ObjectType:   materialize.Schema,
						Name:         schemaName,
						DatabaseName: databaseName,
					}, "materialize_schema_grant.schema_grant", roleName, privilege),
					resource.TestMatchResourceAttr("materialize_schema_grant.schema_grant", "id", terraformGrantIdRegex),
					resource.TestCheckResourceAttr("materialize_schema_grant.schema_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_schema_grant.schema_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_schema_grant.schema_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_schema_grant.schema_grant", "database_name", databaseName),
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

	o := materialize.MaterializeObject{
		ObjectType:   materialize.Schema,
		Name:         schemaName,
		DatabaseName: databaseName,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantSchemaResource(roleName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(o, "materialize_schema_grant.schema_grant", roleName, privilege),
					testAccCheckGrantRevoked(o, roleName, privilege),
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

resource "materialize_schema_grant" "schema_grant" {
	role_name     = materialize_role.test.name
	privilege     = "%s"
	database_name = materialize_database.test.name
	schema_name   = materialize_schema.test.name
}
`, roleName, databaseName, schemaName, privilege)
}

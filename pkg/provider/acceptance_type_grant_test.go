package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					testAccCheckGrantExists(
						materialize.ObjectSchemaStruct{
							ObjectType:   "TYPE",
							Name:         typeName,
							SchemaName:   schemaName,
							DatabaseName: databaseName,
						}, "materialize_type_grant.type_grant", roleName, privilege),
					resource.TestCheckResourceAttr("materialize_type_grant.type_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_type_grant.type_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_type_grant.type_grant", "type_name", typeName),
					resource.TestCheckResourceAttr("materialize_type_grant.type_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_type_grant.type_grant", "database_name", databaseName),
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

	o := materialize.ObjectSchemaStruct{
		ObjectType:   "TYPE",
		Name:         typeName,
		SchemaName:   schemaName,
		DatabaseName: databaseName,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantTypeResource(roleName, typeName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(o, "materialize_type_grant.type_grant", roleName, privilege),
					testAccCheckGrantRevoked(o, roleName, privilege),
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

resource "materialize_type_grant" "type_grant" {
	role_name     = materialize_role.test.name
	privilege     = "%s"
	database_name = materialize_database.test.name
	schema_name   = materialize_schema.test.name
	type_name     = materialize_type.test.name
}
`, roleName, databaseName, schemaName, typeName, privilege)
}

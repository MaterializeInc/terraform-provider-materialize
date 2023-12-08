package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGrantSecret_basic(t *testing.T) {
	privilege := randomPrivilege("SECRET")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantSecretResource(roleName, secretName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(
						materialize.MaterializeObject{
							ObjectType:   "SECRET",
							Name:         secretName,
							SchemaName:   schemaName,
							DatabaseName: databaseName,
						}, "materialize_secret_grant.secret_grant", roleName, privilege),
					resource.TestMatchResourceAttr("materialize_secret_grant.secret_grant", "id", terraformGrantIdRegex),
					resource.TestCheckResourceAttr("materialize_secret_grant.secret_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_secret_grant.secret_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_secret_grant.secret_grant", "secret_name", secretName),
					resource.TestCheckResourceAttr("materialize_secret_grant.secret_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_secret_grant.secret_grant", "database_name", databaseName),
				),
			},
		},
	})
}

func TestAccGrantSecret_disappears(t *testing.T) {
	privilege := randomPrivilege("SECRET")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	o := materialize.MaterializeObject{
		ObjectType:   "SECRET",
		Name:         secretName,
		SchemaName:   schemaName,
		DatabaseName: databaseName,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantSecretResource(roleName, secretName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(o, "materialize_secret_grant.secret_grant", roleName, privilege),
					testAccCheckGrantRevoked(o, roleName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantSecretResource(roleName, secretName, schemaName, databaseName, privilege string) string {
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

resource "materialize_secret" "test" {
	name          = "%s"
	schema_name   = materialize_schema.test.name
	database_name = materialize_database.test.name

	value = "c2VjcmV0Cg=="
}

resource "materialize_secret_grant" "secret_grant" {
	role_name     = materialize_role.test.name
	privilege     = "%s"
	database_name = materialize_database.test.name
	schema_name   = materialize_schema.test.name
	secret_name   = materialize_secret.test.name
}
`, roleName, databaseName, schemaName, secretName, privilege)
}

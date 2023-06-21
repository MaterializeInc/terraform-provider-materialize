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
					testAccCheckGrantSecretExists("materialize_grant_secret.secret_grant", roleName, secretName, schemaName, databaseName, privilege),
					resource.TestCheckResourceAttr("materialize_grant_secret.secret_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_grant_secret.secret_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_grant_secret.secret_grant", "secret_name", secretName),
					resource.TestCheckResourceAttr("materialize_grant_secret.secret_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_grant_secret.secret_grant", "database_name", databaseName),
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantSecretResource(roleName, secretName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantSecretExists("materialize_grant_secret.secret_grant", roleName, secretName, schemaName, databaseName, privilege),
					testAccCheckGrantSecretRevoked(roleName, secretName, schemaName, databaseName, privilege),
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

resource "materialize_grant_secret" "secret_grant" {
	role_name     = materialize_role.test.name
	privilege     = "%s"
	database_name = materialize_database.test.name
	schema_name   = materialize_schema.test.name
	secret_name   = materialize_secret.test.name
}
`, roleName, databaseName, schemaName, secretName, privilege)
}

func testAccCheckGrantSecretExists(grantName, roleName, secretName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("grant not found")
		}

		id, err := materialize.SecretId(db, secretName, schemaName, databaseName)
		if err != nil {
			return err
		}

		roleId, err := materialize.RoleId(db, roleName)
		if err != nil {
			return err
		}

		g, err := materialize.ScanPrivileges(db, "SECRET", id)
		if err != nil {
			return err
		}

		privilegeMap := materialize.ParsePrivileges(g)
		if !materialize.HasPrivilege(privilegeMap[roleId], privilege) {
			return fmt.Errorf("secret object %s does not include privilege %s", g, privilege)
		}
		return nil
	}
}

func testAccCheckGrantSecretRevoked(roleName, secretName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`REVOKE %s ON SECRET "%s"."%s"."%s" FROM "%s";`, privilege, databaseName, schemaName, secretName, roleName))
		return err
	}
}

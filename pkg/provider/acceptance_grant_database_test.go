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

func TestAccGrantDatabase_basic(t *testing.T) {
	privilege := randomPrivilege("DATABASE")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantDatabaseResource(roleName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantDatabaseExists("materialize_grant_database.database_grant", roleName, databaseName, privilege),
					resource.TestCheckResourceAttr("materialize_grant_database.database_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_grant_database.database_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_grant_database.database_grant", "database_name", databaseName),
				),
			},
		},
	})
}

func TestAccGrantDatabase_disappears(t *testing.T) {
	privilege := randomPrivilege("DATABASE")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantDatabaseResource(roleName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantDatabaseExists("materialize_grant_database.database_grant", roleName, databaseName, privilege),
					testAccCheckGrantDatanaseRevoked(roleName, databaseName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantDatabaseResource(roleName, databaseName, privilege string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%s"
}

resource "materialize_database" "test" {
	name = "%s"
}

resource "materialize_grant_database" "database_grant" {
	role_name     = materialize_role.test.name
	privilege     = "%s"
	database_name = materialize_database.test.name
}
`, roleName, databaseName, privilege)
}

func testAccCheckGrantDatabaseExists(grantName, roleName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("grant not found")
		}

		id, err := materialize.DatabaseId(db, databaseName)
		if err != nil {
			return err
		}

		roleId, err := materialize.RoleId(db, roleName)
		if err != nil {
			return err
		}

		g, err := materialize.ScanPrivileges(db, "DATABASE", id)
		if err != nil {
			return err
		}

		priviledgeMap := materialize.ParsePrivileges(g)
		if !materialize.HasPrivilege(priviledgeMap[roleId], privilege) {
			return fmt.Errorf("database object %s does not include privilege %s", g, privilege)
		}
		return nil
	}
}

func testAccCheckGrantDatanaseRevoked(roleName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`REVOKE %s ON DATABASE %s FROM "%s";`, privilege, databaseName, roleName))
		return err
	}
}

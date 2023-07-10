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

func TestAccGrantSecretDefaultPrivilege_basic(t *testing.T) {
	privilege := randomPrivilege("SECRET")
	granteeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	targetName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantSecretDefaultPrivilegeResource(granteeName, targetName, privilege),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("materialize_secret_grant_default_privilege.test", "grantee_name", granteeName),
					resource.TestCheckResourceAttr("materialize_secret_grant_default_privilege.test", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_secret_grant_default_privilege.test", "target_role_name", targetName),
					resource.TestCheckNoResourceAttr("materialize_secret_grant_default_privilege.test", "schema_name"),
					resource.TestCheckNoResourceAttr("materialize_secret_grant_default_privilege.test", "database_name"),
				),
			},
		},
	})
}

func TestAccGrantSecretDefaultPrivilege_disappears(t *testing.T) {
	privilege := randomPrivilege("SECRET")
	granteeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	targetName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantSecretDefaultPrivilegeResource(granteeName, targetName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantSecretDefaultPrivilegeExists("materialize_secret_grant_default_privilege.test", granteeName, targetName, privilege),
					testAccCheckGrantSecretDefaultPrivilegeRevoked(granteeName, targetName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantSecretDefaultPrivilegeResource(granteeName, targetName, privilege string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test_grantee" {
	name = "%[1]s"
}

resource "materialize_role" "test_target" {
	name = "%[2]s"
}

resource "materialize_secret_grant_default_privilege" "test" {
	grantee_name     = materialize_role.test_grantee.name
	privilege        = "%[3]s"
	target_role_name = materialize_role.test_target.name
}
`, granteeName, targetName, privilege)
}

func testAccCheckGrantSecretDefaultPrivilegeExists(grantName, granteeName, targetName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("grant not found")
		}

		granteeId, err := materialize.RoleId(db, grantName)
		if err != nil {
			return err
		}

		targetId, err := materialize.RoleId(db, targetName)
		if err != nil {
			return err
		}

		g, err := materialize.ScanDefaultPrivilege(db, "SECRET", granteeId, targetId, "", "")
		if err != nil {
			return err
		}

		privilegeMap := materialize.ParsePrivileges(g[0].Privileges.String)
		if !materialize.HasPrivilege(privilegeMap[granteeId], privilege) {
			return fmt.Errorf("default privilege %s does not include privilege %s", g[0].Privileges.String, privilege)
		}
		return nil
	}
}

func testAccCheckGrantSecretDefaultPrivilegeRevoked(granteeName, targetName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`ALTER DEFAULT PRIVILEGES FOR ROLE %[1]s REVOKE %[2]s ON SECRETS FROM %[3]s;`, targetName, privilege, granteeName))
		return err
	}
}

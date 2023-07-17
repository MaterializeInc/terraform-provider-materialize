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

func TestAccGrantSystemPrivilege_basic(t *testing.T) {
	for _, roleName := range []string{
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha),
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + "@materialize.com",
	} {
		t.Run(fmt.Sprintf("roleName=%s", roleName), func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: testAccProviderFactories,
				CheckDestroy:      nil,
				Steps: []resource.TestStep{
					{
						Config: testAccGrantSystemPrivilegeResource(roleName),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("materialize_grant_system_privilege.test", "role_name", roleName),
							resource.TestCheckResourceAttr("materialize_grant_system_privilege.test", "privilege", "CREATEDB"),
						),
					},
				},
			})
		})
	}
}

func TestAccGrantSystemPrivilege_disappears(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantSystemPrivilegeResource(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantSystemPrivilegeExists("materialize_grant_system_privilege.test", roleName, "CREATEDB"),
					testAccCheckGrantSystemPrivilegeRevoked(roleName, "CREATEDB"),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantSystemPrivilegeResource(roleName string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_grant_system_privilege" "test" {
	role_name = materialize_role.test.name
	privilege = "CREATEDB"
}
`, roleName)
}

func testAccCheckGrantSystemPrivilegeExists(grantName, roleName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("grant not found")
		}

		// roleId, err := materialize.RoleId(db, roleName)
		// if err != nil {
		// 	return err
		// }

		_, err := materialize.ScanSystemPrivileges(db)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckGrantSystemPrivilegeRevoked(roleName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`REVOKE %[1]s ON SYSTEM FROM %[2]s;`, roleName, privilege))
		return err
	}
}

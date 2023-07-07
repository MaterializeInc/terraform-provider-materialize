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

func TestAccGrantRole_basic(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	granteeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantRoleResource(roleName, granteeName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("materialize_grant_role.test", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_grant_role.test", "member_name", granteeName),
				),
			},
		},
	})
}

func TestAccGrantRole_disappears(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	granteeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantRoleResource(roleName, granteeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantRoleExists("materialize_grant_role.test", roleName, granteeName),
					testAccCheckGrantRoleRevoked(roleName, granteeName),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantRoleResource(roleName, granteeName string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_role" "test_grantee" {
	name = "%[2]s"
}

resource "materialize_grant_role" "test" {
	role_name   = materialize_role.test.name
	member_name = materialize_role.test_grantee.name
}
`, roleName, granteeName)
}

func testAccCheckGrantRoleExists(grantName, roleName, granteeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("grant not found")
		}

		roleId, err := materialize.RoleId(db, roleName)
		if err != nil {
			return err
		}

		granteeId, err := materialize.RoleId(db, granteeName)
		if err != nil {
			return err
		}

		_, err = materialize.ScanRolePrivilege(db, roleId, granteeId)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckGrantRoleRevoked(roleName, granteeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`REVOKE %[1]s FROM %[2]s;`, roleName, granteeName))
		return err
	}
}

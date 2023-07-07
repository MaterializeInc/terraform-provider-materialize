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

func TestAccGrantSystem_basic(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantSystemResource(roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("materialize_grant_system.test", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_grant_system.test", "privilege", "CREATEDB"),
				),
			},
		},
	})
}

func TestAccGrantSystem_disappears(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantSystemResource(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantSystemExists("materialize_grant_system.test", roleName, "CREATEDB"),
					testAccCheckGrantSystemRevoked(roleName, "CREATEDB"),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantSystemResource(roleName string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_grant_system" "test" {
	role_name = materialize_role.test.name
	privilege = "CREATEDB"
}
`, roleName)
}

func testAccCheckGrantSystemExists(grantName, roleName, privilege string) resource.TestCheckFunc {
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

func testAccCheckGrantSystemRevoked(roleName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`REVOKE %[1]s ON SYSTEM FROM %[2]s;`, roleName, privilege))
		return err
	}
}

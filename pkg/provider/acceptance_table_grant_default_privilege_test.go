package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGrantTableDefaultPrivilege_basic(t *testing.T) {
	privilege := randomPrivilege("TABLE")
	granteeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	targetName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantTableDefaultPrivilegeResource(granteeName, targetName, privilege),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test", "grantee_name", granteeName),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test", "target_role_name", targetName),
					resource.TestCheckNoResourceAttr("materialize_table_grant_default_privilege.test", "schema_name"),
					resource.TestCheckNoResourceAttr("materialize_table_grant_default_privilege.test", "database_name"),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test_all", "grantee_name", granteeName),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test_all", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test_all", "target_role_name", "PUBLIC"),
				),
			},
		},
	})
}

func TestAccGrantTableDefaultPrivilege_disappears(t *testing.T) {
	privilege := randomPrivilege("TABLE")
	granteeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	targetName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantTableDefaultPrivilegeResource(granteeName, targetName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantDefaultPrivilegeExists("TABLE", "materialize_table_grant_default_privilege.test", granteeName, targetName, privilege),
					testAccCheckGrantDefaultPrivilegeRevoked("TABLE", granteeName, targetName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantTableDefaultPrivilegeResource(granteeName, targetName, privilege string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test_grantee" {
	name = "%[1]s"
}

resource "materialize_role" "test_target" {
	name = "%[2]s"
}

resource "materialize_table_grant_default_privilege" "test" {
	grantee_name     = materialize_role.test_grantee.name
	privilege        = "%[3]s"
	target_role_name = materialize_role.test_target.name
}

resource "materialize_table_grant_default_privilege" "test_all" {
	grantee_name     = materialize_role.test_grantee.name
	privilege        = "%[3]s"
	target_role_name = "PUBLIC"
}
`, granteeName, targetName, privilege)
}

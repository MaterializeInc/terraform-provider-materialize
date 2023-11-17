package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGrantConnectionDefaultPrivilege_basic(t *testing.T) {
	privilege := randomPrivilege("CONNECTION")
	granteeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	targetName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantConnectionDefaultPrivilegeResource(granteeName, targetName, privilege),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("materialize_connection_grant_default_privilege.test", "grantee_name", granteeName),
					resource.TestCheckResourceAttr("materialize_connection_grant_default_privilege.test", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_connection_grant_default_privilege.test", "target_role_name", targetName),
					resource.TestCheckNoResourceAttr("materialize_connection_grant_default_privilege.test", "schema_name"),
					resource.TestCheckNoResourceAttr("materialize_connection_grant_default_privilege.test", "database_name"),
				),
			},
		},
	})
}

func TestAccGrantConnectionDefaultPrivilege_disappears(t *testing.T) {
	privilege := randomPrivilege("CONNECTION")
	granteeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	targetName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantConnectionDefaultPrivilegeResource(granteeName, targetName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantDefaultPrivilegeExists("CONNECTION", "materialize_connection_grant_default_privilege.test", granteeName, targetName, privilege),
					testAccCheckGrantDefaultPrivilegeRevoked("CONNECTION", granteeName, targetName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantConnectionDefaultPrivilegeResource(granteeName, targetName, privilege string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test_grantee" {
	name = "%[1]s"
}

resource "materialize_role" "test_target" {
	name = "%[2]s"
}

resource "materialize_connection_grant_default_privilege" "test" {
	grantee_name     = materialize_role.test_grantee.name
	privilege        = "%[3]s"
	target_role_name = materialize_role.test_target.name
}
`, granteeName, targetName, privilege)
}

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					testAccCheckGrantDefaultPrivilegeExists("SECRET", "materialize_secret_grant_default_privilege.test", granteeName, targetName, privilege),
					testAccCheckGrantDefaultPrivilegeRevoked("SECRET", granteeName, targetName, privilege),
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

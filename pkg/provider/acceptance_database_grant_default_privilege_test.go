package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGrantDatabaseDefaultPrivilege_basic(t *testing.T) {
	for _, granteeName := range []string{
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha),
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + "@materialize.com",
	} {
		t.Run(fmt.Sprintf("granteeName=%s", granteeName), func(t *testing.T) {
			privilege := randomPrivilege("DATABASE")
			targetName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
			resource.ParallelTest(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: testAccProviderFactories,
				CheckDestroy:      nil,
				Steps: []resource.TestStep{
					{
						Config: testAccGrantDatabaseDefaultPrivilegeResource(granteeName, targetName, privilege),
						Check: resource.ComposeTestCheckFunc(
							resource.TestMatchResourceAttr("materialize_database_grant_default_privilege.test", "id", terraformGrantDefaultIdRegex),
							resource.TestCheckResourceAttr("materialize_database_grant_default_privilege.test", "grantee_name", granteeName),
							resource.TestCheckResourceAttr("materialize_database_grant_default_privilege.test", "privilege", privilege),
							resource.TestCheckResourceAttr("materialize_database_grant_default_privilege.test", "target_role_name", targetName),
							resource.TestCheckNoResourceAttr("materialize_database_grant_default_privilege.test", "database_name"),
						),
					},
				},
			})
		})
	}
}

func TestAccGrantDatabaseDefaultPrivilege_disappears(t *testing.T) {
	privilege := randomPrivilege("DATABASE")
	granteeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	targetName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantDatabaseDefaultPrivilegeResource(granteeName, targetName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantDefaultPrivilegeExists("DATABASE", "materialize_database_grant_default_privilege.test", granteeName, targetName, privilege),
					testAccCheckGrantDefaultPrivilegeRevoked("DATABASE", granteeName, targetName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantDatabaseDefaultPrivilegeResource(granteeName, targetName, privilege string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "test_grantee" {
		name = "%[1]s"
	}

	resource "materialize_role" "test_target" {
		name = "%[2]s"
	}

	resource "materialize_database_grant_default_privilege" "test" {
		grantee_name     = materialize_role.test_grantee.name
		privilege        = "%[3]s"
		target_role_name = materialize_role.test_target.name
	}
	`, granteeName, targetName, privilege)
}

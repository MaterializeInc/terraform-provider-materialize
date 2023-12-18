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
					resource.TestMatchResourceAttr("materialize_table_grant_default_privilege.test", "id", terraformGrantDefaultIdRegex),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test", "grantee_name", granteeName),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test", "target_role_name", targetName),
					resource.TestCheckNoResourceAttr("materialize_table_grant_default_privilege.test", "schema_name"),
					resource.TestCheckNoResourceAttr("materialize_table_grant_default_privilege.test", "database_name"),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test_public_target", "grantee_name", granteeName),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test_public_target", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test_public_target", "target_role_name", "PUBLIC"),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test_public_grantee", "grantee_name", "PUBLIC"),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test_public_grantee", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test_public_grantee", "target_role_name", targetName),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test_public_target_grantee", "grantee_name", "PUBLIC"),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test_public_target_grantee", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_table_grant_default_privilege.test_public_target_grantee", "target_role_name", "PUBLIC"),
				),
				// Deal with non deterministic grants
				Destroy: false,
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

	resource "materialize_table_grant_default_privilege" "test_public_target" {
		grantee_name     = materialize_role.test_grantee.name
		privilege        = "%[3]s"
		target_role_name = "PUBLIC"
	}

	resource "materialize_table_grant_default_privilege" "test_public_grantee" {
		grantee_name     = "PUBLIC"
		privilege        = "%[3]s"
		target_role_name = materialize_role.test_target.name
	}

	resource "materialize_table_grant_default_privilege" "test_public_target_grantee" {
		grantee_name     = "PUBLIC"
		privilege        = "%[3]s"
		target_role_name = "PUBLIC"
	}
	`, granteeName, targetName, privilege)
}

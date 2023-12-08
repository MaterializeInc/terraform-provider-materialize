package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGrantClusterDefaultPrivilege_basic(t *testing.T) {
	privilege := randomPrivilege("CLUSTER")
	granteeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	targetName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantClusterDefaultPrivilegeResource(granteeName, targetName, privilege),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("materialize_cluster_grant_default_privilege.test", "id", terraformGrantDefaultIdRegex),
					resource.TestCheckResourceAttr("materialize_cluster_grant_default_privilege.test", "grantee_name", granteeName),
					resource.TestCheckResourceAttr("materialize_cluster_grant_default_privilege.test", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_cluster_grant_default_privilege.test", "target_role_name", targetName),
				),
			},
		},
	})
}

func TestAccGrantClusterDefaultPrivilege_disappears(t *testing.T) {
	privilege := randomPrivilege("CLUSTER")
	granteeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	targetName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantClusterDefaultPrivilegeResource(granteeName, targetName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantDefaultPrivilegeExists("CLUSTER", "materialize_cluster_grant_default_privilege.test", granteeName, targetName, privilege),
					testAccCheckGrantDefaultPrivilegeRevoked("CLUSTER", granteeName, targetName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantClusterDefaultPrivilegeResource(granteeName, targetName, privilege string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "test_grantee" {
		name = "%[1]s"
	}

	resource "materialize_role" "test_target" {
		name = "%[2]s"
	}

	resource "materialize_cluster_grant_default_privilege" "test" {
		grantee_name     = materialize_role.test_grantee.name
		privilege        = "%[3]s"
		target_role_name = materialize_role.test_target.name
	}
	`, granteeName, targetName, privilege)
}

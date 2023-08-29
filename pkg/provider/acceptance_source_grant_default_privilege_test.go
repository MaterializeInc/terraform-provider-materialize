package provider

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
// 	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
// )

// func TestAccGrantSourceDefaultPrivilege_basic(t *testing.T) {
// 	privilege := randomPrivilege("SOURCE")
// 	granteeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
// 	targetName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:          func() { testAccPreCheck(t) },
// 		ProviderFactories: testAccProviderFactories,
// 		CheckDestroy:      nil,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccGrantSourceDefaultPrivilegeResource(granteeName, targetName, privilege),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr("materialize_source_grant_default_privilege.test", "grantee_name", granteeName),
// 					resource.TestCheckResourceAttr("materialize_source_grant_default_privilege.test", "privilege", privilege),
// 					resource.TestCheckResourceAttr("materialize_source_grant_default_privilege.test", "target_role_name", targetName),
// 					resource.TestCheckNoResourceAttr("materialize_source_grant_default_privilege.test", "schema_name"),
// 					resource.TestCheckNoResourceAttr("materialize_source_grant_default_privilege.test", "database_name"),
// 				),
// 			},
// 		},
// 	})
// }

// func TestAccGrantSourceDefaultPrivilege_disappears(t *testing.T) {
// 	privilege := randomPrivilege("SOURCE")
// 	granteeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
// 	targetName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:          func() { testAccPreCheck(t) },
// 		ProviderFactories: testAccProviderFactories,
// 		CheckDestroy:      nil,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccGrantSourceDefaultPrivilegeResource(granteeName, targetName, privilege),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckGrantDefaultPrivilegeExists("SOURCE", "materialize_source_grant_default_privilege.test", granteeName, targetName, privilege),
// 					testAccCheckGrantDefaultPrivilegeRevoked("SOURCE", granteeName, targetName, privilege),
// 				),
// 				PlanOnly:           true,
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

// func testAccGrantSourceDefaultPrivilegeResource(granteeName, targetName, privilege string) string {
// 	return fmt.Sprintf(`
// resource "materialize_role" "test_grantee" {
// 	name = "%[1]s"
// }

// resource "materialize_role" "test_target" {
// 	name = "%[2]s"
// }

// resource "materialize_source_grant_default_privilege" "test" {
// 	grantee_name     = materialize_role.test_grantee.name
// 	privilege        = "%[3]s"
// 	target_role_name = materialize_role.test_target.name
// }
// `, granteeName, targetName, privilege)
// }

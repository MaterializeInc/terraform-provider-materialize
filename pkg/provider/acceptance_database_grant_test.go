package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGrantDatabase_basic(t *testing.T) {
	for _, roleName := range []string{
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha),
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + "@materialize.com",
	} {
		t.Run(fmt.Sprintf("roleName=%s", roleName), func(t *testing.T) {
			privilege := randomPrivilege("DATABASE")
			databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
			resource.ParallelTest(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: testAccProviderFactories,
				CheckDestroy:      nil,
				Steps: []resource.TestStep{
					{
						Config: testAccGrantDatabaseResource(roleName, databaseName, privilege),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckGrantExists(
								materialize.MaterializeObject{
									ObjectType: "DATABASE",
									Name:       databaseName,
								}, "materialize_database_grant.database_grant", roleName, privilege),
							resource.TestCheckResourceAttr("materialize_database_grant.database_grant", "role_name", roleName),
							resource.TestCheckResourceAttr("materialize_database_grant.database_grant", "privilege", privilege),
							resource.TestCheckResourceAttr("materialize_database_grant.database_grant", "database_name", databaseName),
						),
					},
				},
			})
		})
	}
}

func TestAccGrantDatabase_disappears(t *testing.T) {
	privilege := randomPrivilege("DATABASE")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	o := materialize.MaterializeObject{
		ObjectType: "DATABASE",
		Name:       databaseName,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantDatabaseResource(roleName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(o, "materialize_database_grant.database_grant", roleName, privilege),
					testAccCheckGrantRevoked(o, roleName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantDatabaseResource(roleName, databaseName, privilege string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%s"
}

resource "materialize_database" "test" {
	name = "%s"
}

resource "materialize_database_grant" "database_grant" {
	role_name     = materialize_role.test.name
	privilege     = "%s"
	database_name = materialize_database.test.name
}
`, roleName, databaseName, privilege)
}

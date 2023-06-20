package provider

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jmoiron/sqlx"
)

func TestAccRole_basic(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleResource(roleName, "false", "false", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
					resource.TestCheckResourceAttr("materialize_role.test", "name", roleName),
					resource.TestCheckResourceAttr("materialize_role.test", "inherit", "true"),
					resource.TestCheckResourceAttr("materialize_role.test", "create_role", "false"),
					resource.TestCheckResourceAttr("materialize_role.test", "create_db", "false"),
					resource.TestCheckResourceAttr("materialize_role.test", "create_cluster", "true"),
				),
			},
		},
	})
}

func TestAccRole_update(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleResource(roleName, "false", "false", "true"),
			},
			{
				Config: testAccRoleResource(roleName, "true", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
					resource.TestCheckResourceAttr("materialize_role.test", "name", roleName),
					resource.TestCheckResourceAttr("materialize_role.test", "inherit", "true"),
					resource.TestCheckResourceAttr("materialize_role.test", "create_role", "true"),
					resource.TestCheckResourceAttr("materialize_role.test", "create_db", "false"),
					resource.TestCheckResourceAttr("materialize_role.test", "create_cluster", "false"),
				),
			},
			{
				Config: testAccRoleResource(roleName, "false", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
					resource.TestCheckResourceAttr("materialize_role.test", "name", roleName),
					resource.TestCheckResourceAttr("materialize_role.test", "inherit", "true"),
					resource.TestCheckResourceAttr("materialize_role.test", "create_role", "false"),
					resource.TestCheckResourceAttr("materialize_role.test", "create_db", "false"),
					resource.TestCheckResourceAttr("materialize_role.test", "create_cluster", "false"),
				),
			},
		},
	})
}

func TestAccRole_disappears(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllRolesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleResource(roleName, "true", "true", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
					testAccCheckRoleDisappears(roleName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRoleResource(roleName, createRole, createDB, createCluster string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%s"
	create_role    = %s
	create_db      = %s
	create_cluster = %s
}
`, roleName, createRole, createDB, createCluster)
}

func testAccCheckRoleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("role not found: %s", name)
		}
		_, err := materialize.ScanRole(db, r.Primary.ID)
		return err
	}
}

func testAccCheckRoleDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP ROLE "%s";`, name))
		return err
	}
}

func testAccCheckAllRolesDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_role" {
			continue
		}

		_, err := materialize.ScanRole(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("role %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

package provider

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRole_basic(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleResource(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
					resource.TestMatchResourceAttr("materialize_role.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_role.test", "name", roleName),
					resource.TestCheckResourceAttr("materialize_role.test", "inherit", "true"),
				),
			},
			{
				ResourceName:      "materialize_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRole_update(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	comment := "role comment"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleResource(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
				),
			},
			{
				Config: testAccRoleWithComment(roleName, comment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
					resource.TestCheckResourceAttr("materialize_role.test", "comment", comment),
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
				Config: testAccRoleResource(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "ROLE",
							Name:       roleName,
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRoleResource(roleName string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%s"
}
`, roleName)
}

func testAccRoleWithComment(roleName, comment string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%s"
	comment = "%s"
}
`, roleName, comment)
}

func testAccCheckRoleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("role not found: %s", name)
		}
		_, err = materialize.ScanRole(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllRolesDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_role" {
			continue
		}

		_, err := materialize.ScanRole(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("role %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

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
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRole_withPasswordAndSuperuser(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	password := "secure_password_123"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleWithPasswordAndSuperuser(roleName, password, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
					resource.TestMatchResourceAttr("materialize_role.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_role.test", "name", roleName),
					resource.TestCheckResourceAttr("materialize_role.test", "password", password),
					resource.TestCheckResourceAttr("materialize_role.test", "superuser", "true"),
					resource.TestCheckResourceAttr("materialize_role.test", "inherit", "true"),
				),
			},
			{
				ResourceName:      "materialize_role.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccRole_withLogin(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleWithLogin(roleName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
					resource.TestMatchResourceAttr("materialize_role.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_role.test", "name", roleName),
					resource.TestCheckResourceAttr("materialize_role.test", "login", "true"),
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

func testAccRoleWithPasswordAndSuperuser(roleName, password string, superuser bool) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%s"
	password = "%s"
	superuser = %t
}
`, roleName, password, superuser)
}

func testAccRoleWithLogin(roleName string, login bool) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%s"
	login = %t
}
`, roleName, login)
}

func testAccCheckRoleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
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
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
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

func TestAccRole_withPasswordWo(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	password := "ephemeral_password_value"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleWithPasswordWo(roleName, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
					resource.TestMatchResourceAttr("materialize_role.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_role.test", "name", roleName),
					resource.TestCheckResourceAttr("materialize_role.test", "password_wo_version", "1"),
					resource.TestCheckNoResourceAttr("materialize_role.test", "password_wo"),
				),
			},
			{
				Config: testAccRoleWithPasswordWo(roleName, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
					resource.TestCheckResourceAttr("materialize_role.test", "password_wo_version", "2"),
					resource.TestCheckNoResourceAttr("materialize_role.test", "password_wo"),
				),
			},
			{
				ResourceName:            "materialize_role.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password_wo", "password_wo_version"},
			},
		},
	})
}

func testAccRoleWithPasswordWo(roleName, password string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%s"
	password_wo = "%s"
	password_wo_version = 1
}
`, roleName, password)
}

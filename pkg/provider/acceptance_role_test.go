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
							ObjectType: materialize.Role,
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

func TestAccRole_withLoginAndPassword(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	password := "test_password_456"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleWithLoginAndPassword(roleName, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
					resource.TestMatchResourceAttr("materialize_role.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_role.test", "name", roleName),
					resource.TestCheckResourceAttr("materialize_role.test", "inherit", "true"),
					resource.TestCheckResourceAttr("materialize_role.test", "login", "true"),
					resource.TestCheckResourceAttr("materialize_role.test", "password", password),
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

// TestAccRole_createIfNotExistsAdoptsExisting simulates an externally
// provisioned role (e.g. an SSO/OIDC user whose role is auto-created on first
// login) that exists in Materialize but not in Terraform state, and verifies
// that a materialize_role resource with create_if_not_exists = true adopts it
// instead of failing with "role already exists".
func TestAccRole_createIfNotExistsAdoptsExisting(t *testing.T) {
	seedName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	adoptName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	password := "adopt_password_123"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllRolesDestroyed,
		Steps: []resource.TestStep{
			// Step 1 configures the provider so Meta() is available to the next
			// step's PreConfig.
			{
				Config: testAccRoleResource(seedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
				),
			},
			// Step 2 creates the role out-of-band, then adopts it.
			{
				PreConfig: func() {
					db, _, err := utils.GetDBClientFromMeta(testAccProvider.Meta(), nil)
					if err != nil {
						t.Fatalf("error getting DB client: %s", err)
					}
					o := materialize.MaterializeObject{ObjectType: materialize.Role, Name: adoptName}
					if err := materialize.NewRoleBuilder(db, o).Create(); err != nil {
						t.Fatalf("failed to pre-create role: %s", err)
					}
				},
				Config: testAccRoleCreateIfNotExists(adoptName, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.adopted"),
					resource.TestMatchResourceAttr("materialize_role.adopted", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_role.adopted", "name", adoptName),
					resource.TestCheckResourceAttr("materialize_role.adopted", "create_if_not_exists", "true"),
					resource.TestCheckResourceAttr("materialize_role.adopted", "login", "true"),
					resource.TestCheckResourceAttr("materialize_role.adopted", "password", password),
				),
			},
		},
	})
}

func testAccRoleCreateIfNotExists(roleName, password string) string {
	return fmt.Sprintf(`
resource "materialize_role" "adopted" {
	name                 = "%s"
	login                = true
	password             = "%s"
	create_if_not_exists = true
}
`, roleName, password)
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

func testAccRoleWithLoginAndPassword(roleName, password string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name     = "%s"
	login    = true
	password = "%s"
}
`, roleName, password)
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
				Config: testAccRoleWithPasswordWo(roleName, password, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists("materialize_role.test"),
					resource.TestMatchResourceAttr("materialize_role.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_role.test", "name", roleName),
					resource.TestCheckResourceAttr("materialize_role.test", "password_wo_version", "1"),
					resource.TestCheckNoResourceAttr("materialize_role.test", "password_wo"),
				),
			},
			{
				Config: testAccRoleWithPasswordWo(roleName, password, 2),
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

func testAccRoleWithPasswordWo(roleName, password string, version int) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%s"
	password_wo = "%s"
	password_wo_version = %d
}
`, roleName, password, version)
}

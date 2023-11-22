package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jmoiron/sqlx"
	"golang.org/x/exp/slices"
)

func TestProvider(t *testing.T) {
	if err := Provider("test").InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider("test")
}

var testAccProvider = Provider("test")
var testAccProviderFactories = map[string]func() (*schema.Provider, error){
	"materialize": func() (*schema.Provider, error) { return testAccProvider, nil },
}

func testAccPreCheck(t *testing.T) {

}

func testAccCheckObjectDisappears(object materialize.MaterializeObject) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP %[1]s %[2]s;`, object.ObjectType, object.QualifiedName()))
		return err
	}
}

func testAccCheckGrantRevoked(object materialize.MaterializeObject, roleName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(
			`REVOKE %[1]s ON %[2]s %[3]s FROM "%[4]s";`,
			privilege, object.ObjectType, object.QualifiedName(), roleName,
		))
		return err
	}
}

func testAccCheckGrantExists(object materialize.MaterializeObject, grantName, roleName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("grant not found")
		}
		id, err := materialize.ObjectId(db, object)
		if err != nil {
			return err
		}
		roleId, err := materialize.RoleId(db, roleName)
		if err != nil {
			return err
		}
		g, err := materialize.ScanPrivileges(db, object.ObjectType, id)
		if err != nil {
			return err
		}
		p, _ := materialize.MapGrantPrivileges(g)
		if !slices.Contains(p[roleId], privilege) {
			return fmt.Errorf("object %s does not include privilege %s", p, privilege)
		}
		return nil
	}
}

func testAccCheckGrantDefaultPrivilegeRevoked(objectType, granteeName, targetName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`ALTER DEFAULT PRIVILEGES FOR ROLE %[1]s REVOKE %[2]s ON %[3]sS FROM %[4]s;`, targetName, privilege, objectType, granteeName))
		return err
	}
}

func testAccCheckGrantDefaultPrivilegeExists(objectType, grantName, granteeName, targetName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("default grant not found")
		}
		granteeId, err := materialize.RoleId(db, grantName)
		if err != nil {
			return err
		}
		targetId, err := materialize.RoleId(db, targetName)
		if err != nil {
			return err
		}
		g, err := materialize.ScanDefaultPrivilege(db, objectType, granteeId, targetId, "", "")
		if err != nil {
			return err
		}
		p, _ := materialize.MapDefaultGrantPrivileges(g)
		if !slices.Contains(p[granteeId], privilege) {
			return fmt.Errorf("object %s does not include privilege %s", p, privilege)
		}
		return nil
	}
}

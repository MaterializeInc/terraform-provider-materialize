package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"golang.org/x/exp/slices"
)

var (
	terraformObjectIdRegex       = regexp.MustCompile("^aws/us-east-1:")
	terraformObjectTypeIdRegex   = regexp.MustCompile("^aws/us-east-1:id:")
	terraformGrantIdRegex        = regexp.MustCompile("^aws/us-east-1:GRANT|")
	terraformGrantDefaultIdRegex = regexp.MustCompile("^aws/us-east-1:GRANT DEFAULT|")
	terraformGrantSystemIdRegex  = regexp.MustCompile("^aws/us-east-1:GRANT ROLE|")
	terraformGrantRoleIdRegex    = regexp.MustCompile("^aws/us-east-1:GRANT SYSTEM|")
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
	if os.Getenv("MZ_ENDPOINT") == "" {
		t.Fatal("MZ_ENDPOINT must be set for acceptance tests")
	}
	if os.Getenv("MZ_PASSWORD") == "" {
		t.Fatal("MZ_PASSWORD must be set for acceptance tests")
	}
	if os.Getenv("MZ_CLOUD_ENDPOINT") == "" {
		t.Fatal("MZ_CLOUD_ENDPOINT must be set for acceptance tests")
	}
}

func testAccAddColumnComment(object materialize.MaterializeObject, column, comment string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		_, err = db.Exec(fmt.Sprintf(`COMMENT ON COLUMN %[1]s.%[2]s IS %[3]s;`,
			object.QualifiedName(),
			column,
			materialize.QuoteString(comment),
		))
		return err
	}
}

func testAccCheckObjectDisappears(object materialize.MaterializeObject) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		_, err = db.Exec(fmt.Sprintf(`DROP %[1]s %[2]s;`, object.ObjectType, object.QualifiedName()))
		return err
	}
}

func testAccCheckGrantRevoked(object materialize.MaterializeObject, roleName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		_, err = db.Exec(fmt.Sprintf(
			`REVOKE %[1]s ON %[2]s %[3]s FROM "%[4]s";`,
			privilege, object.ObjectType, object.QualifiedName(), roleName,
		))
		return err
	}
}

func testAccCheckGrantExists(object materialize.MaterializeObject, grantName, roleName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
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
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		_, err = db.Exec(fmt.Sprintf(`ALTER DEFAULT PRIVILEGES FOR ROLE %[1]s REVOKE %[2]s ON %[3]sS FROM %[4]s;`, targetName, privilege, objectType, granteeName))
		return err
	}
}

func testAccCheckGrantDefaultPrivilegeExists(objectType, grantName, granteeName, targetName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
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

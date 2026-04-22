package provider

import (
	"fmt"
	"os"
	"regexp"
	"strings"
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

func TestProviderOptionsSchema(t *testing.T) {
	p := Provider("test")
	s, ok := p.Schema["options"]
	if !ok {
		t.Fatal("provider schema missing `options` field")
	}
	if s.Type != schema.TypeMap {
		t.Fatalf("expected options Type == TypeMap, got %v", s.Type)
	}
	if !s.Optional {
		t.Fatal("expected options to be Optional")
	}
	elem, ok := s.Elem.(*schema.Schema)
	if !ok {
		t.Fatalf("expected options Elem to be *schema.Schema, got %T", s.Elem)
	}
	if elem.Type != schema.TypeString {
		t.Fatalf("expected options Elem Type == TypeString, got %v", elem.Type)
	}
	if s.ValidateDiagFunc == nil {
		t.Fatal("expected options to have a ValidateDiagFunc")
	}
}

func TestValidateProviderOptions(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
		errSub  string
	}{
		{
			name:  "valid keys pass",
			input: map[string]interface{}{"cluster": "quickstart", "search_path": "public"},
		},
		{
			name:  "nil input",
			input: nil,
		},
		{
			name:  "empty map",
			input: map[string]interface{}{},
		},
		{
			name:  "oidc option passes",
			input: map[string]interface{}{"oidc_auth_enabled": "true"},
		},
		{
			name:    "transaction_isolation is reserved",
			input:   map[string]interface{}{"transaction_isolation": "serializable"},
			wantErr: true,
			errSub:  "transaction_isolation",
		},
		{
			name:    "application_name is reserved",
			input:   map[string]interface{}{"application_name": "custom"},
			wantErr: true,
			errSub:  "application_name",
		},
		{
			name:    "reserved key check is case-insensitive",
			input:   map[string]interface{}{"Transaction_Isolation": "serializable"},
			wantErr: true,
			errSub:  "Transaction_Isolation",
		},
		{
			name:    "reserved key check catches uppercase application_name",
			input:   map[string]interface{}{"APPLICATION_NAME": "custom"},
			wantErr: true,
			errSub:  "APPLICATION_NAME",
		},
		{
			name:    "invalid key with space",
			input:   map[string]interface{}{"bad key": "v"},
			wantErr: true,
			errSub:  "invalid option key",
		},
		{
			name:    "invalid key starting with digit",
			input:   map[string]interface{}{"1abc": "v"},
			wantErr: true,
			errSub:  "invalid option key",
		},
		{
			name:    "empty key",
			input:   map[string]interface{}{"": "v"},
			wantErr: true,
			errSub:  "invalid option key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diags := validateProviderOptions(tc.input, nil)
			hasErr := diags.HasError()
			if hasErr != tc.wantErr {
				t.Fatalf("got diags=%v, wantErr=%v", diags, tc.wantErr)
			}
			if tc.wantErr && tc.errSub != "" {
				found := false
				for _, d := range diags {
					if strings.Contains(d.Summary, tc.errSub) || strings.Contains(d.Detail, tc.errSub) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected a diagnostic mentioning %q, got %v", tc.errSub, diags)
				}
			}
		})
	}
}

func TestOptionsFromResourceData(t *testing.T) {
	s := Provider("test").Schema
	r := schema.TestResourceDataRaw(t, s, map[string]interface{}{
		"options": map[string]interface{}{
			"cluster":           "quickstart",
			"oidc_auth_enabled": "true",
		},
	})
	got := optionsFromResourceData(r)
	if got["cluster"] != "quickstart" || got["oidc_auth_enabled"] != "true" {
		t.Fatalf("unexpected options map: %v", got)
	}

	empty := schema.TestResourceDataRaw(t, s, map[string]interface{}{})
	if optionsFromResourceData(empty) != nil {
		t.Fatal("expected nil options when map is unset")
	}
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

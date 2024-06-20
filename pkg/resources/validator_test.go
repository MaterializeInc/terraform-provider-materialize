package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type testCase struct {
	val  interface{}
	f    schema.SchemaValidateFunc
	pass bool
}

func runTestCases(t *testing.T, cases []testCase) {
	t.Helper()

	for i, tc := range cases {
		_, errs := tc.f(tc.val, "test_property")

		if len(errs) == 0 && tc.pass {
			continue
		}

		if len(errs) != 0 && tc.pass {
			t.Fatalf("expected test case %d to produce no errors, got %v", i, errs)
		}
	}
}

func TestValidPrivileges(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val:  "SELECT",
			f:    validPrivileges("VIEW"),
			pass: true,
		},
		{
			val:  "CREATE",
			f:    validPrivileges("DATABASE"),
			pass: true,
		},
		{
			val:  "SELECT",
			f:    validPrivileges("DATABASE"),
			pass: false,
		},
	})
}

func TestValidateServiceUsername(t *testing.T) {
	tests := []struct {
		name     string
		val      interface{}
		key      string
		wantWarn []string
		wantErr  string
	}{
		{
			name:    "Valid username",
			val:     "username",
			key:     "user",
			wantErr: "",
		},
		{
			name:    "Username with @",
			val:     "user@name",
			key:     "user",
			wantErr: `"user" must not contain '@', got: user@name`,
		},
		{
			name:    "Username with forbidden prefix mz_",
			val:     "mz_username",
			key:     "user",
			wantErr: `"user" must not start with mz_, got: mz_username`,
		},
		{
			name:    "Username with forbidden prefix pg_",
			val:     "pg_username",
			key:     "user",
			wantErr: `"user" must not start with pg_, got: pg_username`,
		},
		{
			name:    "Username with forbidden prefix external_",
			val:     "external_username",
			key:     "user",
			wantErr: `"user" must not start with external_, got: external_username`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateServiceUsername(tt.val, tt.key)
			if (len(errs) == 0 && tt.wantErr != "") || (len(errs) > 0 && errs[0].Error() != tt.wantErr) {
				t.Errorf("validateServiceUsername() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

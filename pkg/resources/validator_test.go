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

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleParameter_basic(t *testing.T) {
	resourceRoleName := "materialize_role.test_role"
	resourceRoleParameterName := "materialize_role_parameter.test"
	roleName := fmt.Sprintf("test_role_%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	variableName := "transaction_isolation"
	variableValue := "read committed"
	variableValueUpdated := "repeatable read"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleParameterConfig(roleName, variableName, variableValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRoleName, "name", roleName),
					resource.TestCheckResourceAttr(resourceRoleParameterName, "role_name", roleName),
					resource.TestCheckResourceAttr(resourceRoleParameterName, "variable_name", variableName),
					resource.TestCheckResourceAttr(resourceRoleParameterName, "variable_value", variableValue),
				),
			},
			{
				Config: testAccRoleParameterConfig(roleName, variableName, variableValueUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceRoleName, "name", roleName),
					resource.TestCheckResourceAttr(resourceRoleParameterName, "role_name", roleName),
					resource.TestCheckResourceAttr(resourceRoleParameterName, "variable_name", variableName),
					resource.TestCheckResourceAttr(resourceRoleParameterName, "variable_value", variableValueUpdated),
				),
			},
		},
	})
}

func testAccRoleParameterConfig(roleName, variableName, variableValue string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test_role" {
  name = "%s"
}

resource "materialize_role_parameter" "test" {
  role_name      = materialize_role.test_role.name
  variable_name  = "%s"
  variable_value = "%s"
}
`, roleName, variableName, variableValue)
}

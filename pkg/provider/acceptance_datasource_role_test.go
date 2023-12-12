package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceRole_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRole(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					// Cannot ensure the exact number of objects with parallel tests
					// Ensuring minimum
					resource.TestMatchResourceAttr("data.materialize_role.test_all", "roles.#", regexp.MustCompile("([4-9]|\\d{2,})")),
				),
			},
		},
	})
}

func testAccDatasourceRole(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "a" {
		name         = "%[1]s_a"
	}

	resource "materialize_role" "b" {
		name         = "%[1]s_b"
	}

	resource "materialize_role" "c" {
		name         = "%[1]s_c"
	}

	resource "materialize_role" "d" {
		name         = "%[1]s_d"
	}

	data "materialize_role" "test_all" {
		depends_on    = [
			materialize_role.a,
			materialize_role.b,
			materialize_role.c,
			materialize_role.d,
		]
	}
	`, nameSpace)
}

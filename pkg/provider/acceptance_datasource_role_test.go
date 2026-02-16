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

func TestAccDatasourceRole_withPattern(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRoleWithPattern(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					// Should match exactly 2 roles with pattern
					resource.TestCheckResourceAttr("data.materialize_role.test_pattern", "roles.#", "2"),
					// Verify the pattern filter is set
					resource.TestCheckResourceAttr("data.materialize_role.test_pattern", "like_pattern", fmt.Sprintf("%s_b%%", nameSpace)),
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

func testAccDatasourceRoleWithPattern(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "b1" {
		name         = "%[1]s_b1"
	}

	resource "materialize_role" "b2" {
		name         = "%[1]s_b2"
	}

	resource "materialize_role" "c" {
		name         = "%[1]s_c"
	}

	data "materialize_role" "test_pattern" {
		like_pattern = "%[1]s_b%%"
		depends_on   = [
			materialize_role.b1,
			materialize_role.b2,
			materialize_role.c,
		]
	}
	`, nameSpace)
}

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceDatabase_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDatabase(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					// Cannot ensure the exact number of objects with parallel tests
					// Ensuring minimum
					resource.TestMatchResourceAttr("data.materialize_database.test_all", "databases.#", regexp.MustCompile("([3-9]|\\d{2,})")),
				),
			},
		},
	})
}

func testAccDatasourceDatabase(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_database" "a" {
		name    = "%[1]s_a"
	}

	resource "materialize_database" "b" {
		name    = "%[1]s_b"
	}

	resource "materialize_database" "c" {
		name    = "%[1]s_c"
	}

	data "materialize_database" "test_all" {
		depends_on    = [
			materialize_database.a,
			materialize_database.b,
			materialize_database.c,
		]
	}
	`, nameSpace)
}

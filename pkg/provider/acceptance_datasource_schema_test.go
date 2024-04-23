package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceSchema_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceSchema(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.materialize_schema.test_database", "database_name", nameSpace),
					// Will be one greater than defined for public schema
					resource.TestCheckResourceAttr("data.materialize_schema.test_database", "schemas.#", "2"),
					resource.TestCheckResourceAttr("data.materialize_schema.test_database_2", "database_name", nameSpace+"_2"),
					resource.TestCheckResourceAttr("data.materialize_schema.test_database_2", "schemas.#", "3"),
					resource.TestCheckNoResourceAttr("data.materialize_schema.test_all", "database_name"),
					// Cannot ensure the exact number of objects with parallel tests
					// Ensuring minimum
					resource.TestMatchResourceAttr("data.materialize_schema.test_all", "schemas.#", regexp.MustCompile("([6-9]|\\d{2,})")),
				),
			},
		},
	})
}

func testAccDatasourceSchema(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_database" "test" {
		name    = "%[1]s"
	}

	resource "materialize_database" "test_2" {
		name    = "%[1]s_2"
	}

	resource "materialize_schema" "a" {
		name          = "%[1]s_a"
		database_name = materialize_database.test.name
	}

	resource "materialize_schema" "b" {
		name          = "%[1]s_b"
		database_name = materialize_database.test.name
	}

	resource "materialize_schema" "c" {
		name          = "%[1]s_c"
		database_name = materialize_database.test_2.name
	}

	resource "materialize_schema" "d" {
		name          = "%[1]s_d"
		database_name = materialize_database.test_2.name
	}

	resource "materialize_schema" "e" {
		name          = "%[1]s_e"
		database_name = materialize_database.test_2.name
	}

	data "materialize_schema" "test_all" {
		depends_on    = [
			materialize_schema.a,
			materialize_schema.b,
			materialize_schema.c,
			materialize_schema.d,
			materialize_schema.e,
		]
	}

	data "materialize_schema" "test_database" {
		database_name = materialize_database.test.name
		depends_on    = [
			materialize_schema.a,
			materialize_schema.b,
			materialize_schema.c,
			materialize_schema.d,
			materialize_schema.e,
		]
	}

	data "materialize_schema" "test_database_2" {
		database_name = materialize_database.test_2.name
		depends_on = [
			materialize_schema.a,
			materialize_schema.b,
			materialize_schema.c,
			materialize_schema.d,
			materialize_schema.e,
		]
	}
	`, nameSpace)
}

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceView_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceView(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.materialize_view.test_database", "database_name", nameSpace),
					resource.TestCheckNoResourceAttr("data.materialize_view.test_database", "schema_name"),
					resource.TestCheckResourceAttr("data.materialize_view.test_database", "views.#", "3"),
					resource.TestCheckResourceAttr("data.materialize_view.test_database_schema", "database_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_view.test_database_schema", "schema_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_view.test_database_schema", "views.#", "2"),
					resource.TestCheckResourceAttr("data.materialize_view.test_database_2", "database_name", nameSpace+"_2"),
					resource.TestCheckNoResourceAttr("data.materialize_view.test_database_2", "schema_name"),
					resource.TestCheckResourceAttr("data.materialize_view.test_database_2", "views.#", "2"),
					resource.TestCheckNoResourceAttr("data.materialize_view.test_all", "database_name"),
					resource.TestCheckNoResourceAttr("data.materialize_view.test_all", "schema_name"),
					// Cannot ensure the exact number of objects with parallel tests
					// Ensuring minimum
					resource.TestMatchResourceAttr("data.materialize_view.test_all", "views.#", regexp.MustCompile("([5-9]|\\d{2,})")),
				),
			},
		},
	})
}

func testAccDatasourceView(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_database" "test" {
		name    = "%[1]s"
	}

	resource "materialize_database" "test_2" {
		name    = "%[1]s_2"
	}

	resource "materialize_schema" "test" {
		name          = "%[1]s"
		database_name = materialize_database.test.name
	}

	resource "materialize_view" "a" {
		name          = "%[1]s_a"
		database_name = materialize_database.test.name
  		statement = <<SQL
			SELECT
    		1 AS id
		SQL
	}

	resource "materialize_view" "b" {
		name          = "%[1]s_b"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
  		statement = <<SQL
			SELECT
    		1 AS id
		SQL
	}

	resource "materialize_view" "c" {
		name          = "%[1]s_c"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
  		statement = <<SQL
			SELECT
    		1 AS id
		SQL
	}

	resource "materialize_view" "d" {
		name          = "%[1]s_d"
		database_name = materialize_database.test_2.name
  		statement = <<SQL
			SELECT
    		1 AS id
		SQL
	}

	resource "materialize_view" "e" {
		name          = "%[1]s_e"
		database_name = materialize_database.test_2.name
  		statement = <<SQL
			SELECT
    		1 AS id
		SQL
	}

	data "materialize_view" "test_all" {
		depends_on    = [
			materialize_view.a,
			materialize_view.b,
			materialize_view.c,
			materialize_view.d,
			materialize_view.e,
		]
	}

	data "materialize_view" "test_database" {
		database_name = materialize_database.test.name
		depends_on    = [
			materialize_view.a,
			materialize_view.b,
			materialize_view.c,
			materialize_view.d,
			materialize_view.e,
		]
	}
	
	data "materialize_view" "test_database_schema" {
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		depends_on    = [
			materialize_view.a,
			materialize_view.b,
			materialize_view.c,
			materialize_view.d,
			materialize_view.e,
		]
	}

	data "materialize_view" "test_database_2" {
		database_name = materialize_database.test_2.name
		depends_on = [
			materialize_view.a,
			materialize_view.b,
			materialize_view.c,
			materialize_view.d,
			materialize_view.e,
		]
	}
	`, nameSpace)
}

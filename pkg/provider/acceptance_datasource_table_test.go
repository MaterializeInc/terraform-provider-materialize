package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceTable_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceTable(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.materialize_table.test_database", "database_name", nameSpace),
					resource.TestCheckNoResourceAttr("data.materialize_table.test_database", "schema_name"),
					resource.TestCheckResourceAttr("data.materialize_table.test_database", "tables.#", "3"),
					resource.TestCheckResourceAttr("data.materialize_table.test_database_schema", "database_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_table.test_database_schema", "schema_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_table.test_database_schema", "tables.#", "3"),
					resource.TestCheckResourceAttr("data.materialize_table.test_database_2", "database_name", nameSpace+"_2"),
					resource.TestCheckNoResourceAttr("data.materialize_table.test_database_2", "schema_name"),
					resource.TestCheckResourceAttr("data.materialize_table.test_database_2", "tables.#", "2"),
					resource.TestCheckNoResourceAttr("data.materialize_table.test_all", "database_name"),
					resource.TestCheckNoResourceAttr("data.materialize_table.test_all", "schema_name"),
					// Cannot ensure the exact number of objects with parallel tests
					// Ensuring minimum
					resource.TestMatchResourceAttr("data.materialize_table.test_all", "tables.#", regexp.MustCompile("([5-9]|\\d{2,})")),
				),
			},
		},
	})
}

func testAccDatasourceTable(nameSpace string) string {
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

	resource "materialize_schema" "test_2" {
		name          = "%[1]s_2"
		database_name = materialize_database.test_2.name
	}


	resource "materialize_table" "a" {
		name          = "%[1]s_a"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		comment       = "some comment"

		column {
			name    = "column_1"
			type    = "text"
			comment = "some comment"
		}
	}

	resource "materialize_table" "b" {
		name          = "%[1]s_b"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		column {
			name = "column_1"
			type = "text"
		}
	}

	resource "materialize_table" "c" {
		name          = "%[1]s_c"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		comment       = "some comment"

		column {
			name    = "column_1"
			type    = "text"
			comment = "some comment"
		}
	}

	resource "materialize_table" "d" {
		name          = "%[1]s_d"
		database_name = materialize_database.test_2.name
		schema_name   = materialize_schema.test_2.name
		column {
			name = "column_1"
			type = "text"
		}
	}

	resource "materialize_table" "e" {
		name          = "%[1]s_e"
		database_name = materialize_database.test_2.name
		schema_name   = materialize_schema.test_2.name
		column {
			name = "column_1"
			type = "text"
		}
	}

	data "materialize_table" "test_all" {
		depends_on    = [
			materialize_table.a,
			materialize_table.b,
			materialize_table.c,
			materialize_table.d,
			materialize_table.e,
		]
	}

	data "materialize_table" "test_database" {
		database_name = materialize_database.test.name
		depends_on    = [
			materialize_table.a,
			materialize_table.b,
			materialize_table.c,
			materialize_table.d,
			materialize_table.e,
		]
	}
	
	data "materialize_table" "test_database_schema" {
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		depends_on    = [
			materialize_table.a,
			materialize_table.b,
			materialize_table.c,
			materialize_table.d,
			materialize_table.e,
		]
	}

	data "materialize_table" "test_database_2" {
		database_name = materialize_database.test_2.name
		depends_on = [
			materialize_table.a,
			materialize_table.b,
			materialize_table.c,
			materialize_table.d,
			materialize_table.e,
		]
	}
	`, nameSpace)
}

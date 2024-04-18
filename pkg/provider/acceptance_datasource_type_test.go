package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceType_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceType(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.materialize_type.test_database", "database_name", nameSpace),
					resource.TestCheckNoResourceAttr("data.materialize_type.test_database", "schema_name"),
					resource.TestCheckResourceAttr("data.materialize_type.test_database", "types.#", "3"),
					resource.TestCheckResourceAttr("data.materialize_type.test_database_schema", "database_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_type.test_database_schema", "schema_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_type.test_database_schema", "types.#", "2"),
					resource.TestCheckResourceAttr("data.materialize_type.test_database_2", "database_name", nameSpace+"_2"),
					resource.TestCheckNoResourceAttr("data.materialize_type.test_database_2", "schema_name"),
					resource.TestCheckResourceAttr("data.materialize_type.test_database_2", "types.#", "2"),
					resource.TestCheckNoResourceAttr("data.materialize_type.test_all", "database_name"),
					resource.TestCheckNoResourceAttr("data.materialize_type.test_all", "schema_name"),
					// Cannot ensure the exact number of objects with parallel tests
					// Ensuring minimum
					resource.TestMatchResourceAttr("data.materialize_type.test_all", "types.#", regexp.MustCompile("([5-9]|\\d{2,})")),
				),
			},
		},
	})
}

func testAccDatasourceType(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_database" "test" {
		name    = "%[1]s"
	}

	resource "materialize_database" "test_2" {
		name    = "%[1]s_2"
	}

	resource "materialize_schema" "public_schema" {
		name          = "public"
		database_name = materialize_database.test.name
	}

	resource "materialize_schema" "public_schema2" {
		name          = "public"
		database_name = materialize_database.test_2.name
	}

	resource "materialize_schema" "test" {
		name          = "%[1]s"
		database_name = materialize_database.test.name
	}

	resource "materialize_type" "a" {
		name          = "%[1]s_a"
		database_name = materialize_database.test.name
		list_properties {
			element_type = "int4"
  		}
	}

	resource "materialize_type" "b" {
		name          = "%[1]s_b"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		list_properties {
			element_type = "int4"
  		}
	}

	resource "materialize_type" "c" {
		name          = "%[1]s_c"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		list_properties {
			element_type = "int4"
  		}
	}

	resource "materialize_type" "d" {
		name          = "%[1]s_d"
		database_name = materialize_database.test_2.name
		list_properties {
			element_type = "int4"
  		}
	}

	resource "materialize_type" "e" {
		name          = "%[1]s_e"
		database_name = materialize_database.test_2.name
		list_properties {
			element_type = "int4"
  		}
	}

	data "materialize_type" "test_all" {
		depends_on    = [
			materialize_type.a,
			materialize_type.b,
			materialize_type.c,
			materialize_type.d,
			materialize_type.e,
		]
	}

	data "materialize_type" "test_database" {
		database_name = materialize_database.test.name
		depends_on    = [
			materialize_type.a,
			materialize_type.b,
			materialize_type.c,
			materialize_type.d,
			materialize_type.e,
		]
	}
	
	data "materialize_type" "test_database_schema" {
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		depends_on    = [
			materialize_type.a,
			materialize_type.b,
			materialize_type.c,
			materialize_type.d,
			materialize_type.e,
		]
	}

	data "materialize_type" "test_database_2" {
		database_name = materialize_database.test_2.name
		depends_on = [
			materialize_type.a,
			materialize_type.b,
			materialize_type.c,
			materialize_type.d,
			materialize_type.e,
		]
	}
	`, nameSpace)
}

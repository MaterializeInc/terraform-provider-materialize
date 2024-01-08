package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceSource_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceSource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.materialize_source.test_database", "database_name", nameSpace),
					resource.TestCheckNoResourceAttr("data.materialize_source.test_database", "schema_name"),
					// Will be double the amount for subsources
					resource.TestCheckResourceAttr("data.materialize_source.test_database", "sources.#", "6"),
					resource.TestCheckResourceAttr("data.materialize_source.test_database_schema", "database_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_source.test_database_schema", "schema_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_source.test_database_schema", "sources.#", "4"),
					resource.TestCheckResourceAttr("data.materialize_source.test_database_2", "database_name", nameSpace+"_2"),
					resource.TestCheckNoResourceAttr("data.materialize_source.test_database_2", "schema_name"),
					resource.TestCheckResourceAttr("data.materialize_source.test_database_2", "sources.#", "4"),
					resource.TestCheckNoResourceAttr("data.materialize_source.test_all", "database_name"),
					resource.TestCheckNoResourceAttr("data.materialize_source.test_all", "schema_name"),
					// Cannot ensure the exact number of objects with parallel tests
					// Ensuring minimum
					resource.TestMatchResourceAttr("data.materialize_source.test_all", "sources.#", regexp.MustCompile("([9]|\\d{2,})")),
				),
			},
		},
	})
}

func testAccDatasourceSource(nameSpace string) string {
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

	resource "materialize_source_load_generator" "a" {
		name          = "%[1]s_a"
		database_name = materialize_database.test.name
		cluster_name  = "quickstart"
		load_generator_type = "COUNTER"
	}

	resource "materialize_source_load_generator" "b" {
		name          = "%[1]s_b"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		cluster_name  = "quickstart"
		load_generator_type = "COUNTER"
	}

	resource "materialize_source_load_generator" "c" {
		name          = "%[1]s_c"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		cluster_name  = "quickstart"
		load_generator_type = "COUNTER"
	}

	resource "materialize_source_load_generator" "d" {
		name          = "%[1]s_d"
		database_name = materialize_database.test_2.name
		cluster_name  = "quickstart"
		load_generator_type = "COUNTER"
	}

	resource "materialize_source_load_generator" "e" {
		name          = "%[1]s_e"
		database_name = materialize_database.test_2.name
		cluster_name  = "quickstart"
		load_generator_type = "COUNTER"
	}

	data "materialize_source" "test_all" {
		depends_on    = [
			materialize_source_load_generator.a,
			materialize_source_load_generator.b,
			materialize_source_load_generator.c,
			materialize_source_load_generator.d,
			materialize_source_load_generator.e,
		]
	}

	data "materialize_source" "test_database" {
		database_name = materialize_database.test.name
		depends_on    = [
			materialize_source_load_generator.a,
			materialize_source_load_generator.b,
			materialize_source_load_generator.c,
			materialize_source_load_generator.d,
			materialize_source_load_generator.e,
		]
	}
	
	data "materialize_source" "test_database_schema" {
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		depends_on    = [
			materialize_source_load_generator.a,
			materialize_source_load_generator.b,
			materialize_source_load_generator.c,
			materialize_source_load_generator.d,
			materialize_source_load_generator.e,
		]
	}

	data "materialize_source" "test_database_2" {
		database_name = materialize_database.test_2.name
		depends_on = [
			materialize_source_load_generator.a,
			materialize_source_load_generator.b,
			materialize_source_load_generator.c,
			materialize_source_load_generator.d,
			materialize_source_load_generator.e,
		]
	}
	`, nameSpace)
}

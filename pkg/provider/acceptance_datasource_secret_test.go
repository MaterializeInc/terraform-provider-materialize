package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceSecret_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceSecret(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.materialize_secret.test_database", "database_name", nameSpace),
					resource.TestCheckNoResourceAttr("data.materialize_secret.test_database", "schema_name"),
					resource.TestCheckResourceAttr("data.materialize_secret.test_database", "secrets.#", "3"),
					resource.TestCheckResourceAttr("data.materialize_secret.test_database_schema", "database_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_secret.test_database_schema", "schema_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_secret.test_database_schema", "secrets.#", "3"),
					resource.TestCheckResourceAttr("data.materialize_secret.test_database_2", "database_name", nameSpace+"_2"),
					resource.TestCheckNoResourceAttr("data.materialize_secret.test_database_2", "schema_name"),
					resource.TestCheckResourceAttr("data.materialize_secret.test_database_2", "secrets.#", "2"),
					resource.TestCheckNoResourceAttr("data.materialize_secret.test_all", "database_name"),
					resource.TestCheckNoResourceAttr("data.materialize_secret.test_all", "schema_name"),
					// Cannot ensure the exact number of objects with parallel tests
					// Ensuring minimum
					resource.TestMatchResourceAttr("data.materialize_secret.test_all", "secrets.#", regexp.MustCompile("([5-9]|\\d{2,})")),
				),
			},
		},
	})
}

func testAccDatasourceSecret(nameSpace string) string {
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

	resource "materialize_secret" "a" {
		name          = "%[1]s_a"
		value         = "some-secret-value"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
	}

	resource "materialize_secret" "b" {
		name          = "%[1]s_b"
		value         = "some-secret-value"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
	}

	resource "materialize_secret" "c" {
		name          = "%[1]s_c"
		value         = "some-secret-value"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
	}

	resource "materialize_secret" "d" {
		name          = "%[1]s_d"
		value         = "some-secret-value"
		database_name = materialize_database.test_2.name
		schema_name   = materialize_schema.test_2.name
	}

	resource "materialize_secret" "e" {
		name  = "%[1]s_e"
		value = "some-secret-value"
		database_name = materialize_database.test_2.name
		schema_name   = materialize_schema.test_2.name
	}

	data "materialize_secret" "test_all" {
		depends_on    = [
			materialize_secret.a,
			materialize_secret.b,
			materialize_secret.c,
			materialize_secret.d,
			materialize_secret.e,
		]
	}

	data "materialize_secret" "test_database" {
		database_name = materialize_database.test.name
		depends_on    = [
			materialize_secret.a,
			materialize_secret.b,
			materialize_secret.c,
			materialize_secret.d,
			materialize_secret.e,
		]
	}
	
	data "materialize_secret" "test_database_schema" {
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		depends_on    = [
			materialize_secret.a,
			materialize_secret.b,
			materialize_secret.c,
			materialize_secret.d,
			materialize_secret.e,
		]
	}

	data "materialize_secret" "test_database_2" {
		database_name = materialize_database.test_2.name
		depends_on = [
			materialize_secret.a,
			materialize_secret.b,
			materialize_secret.c,
			materialize_secret.d,
			materialize_secret.e,
		]
	}
	`, nameSpace)
}

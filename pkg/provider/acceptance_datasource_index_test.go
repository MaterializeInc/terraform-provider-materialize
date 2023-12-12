package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceIndex_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceIndex(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.materialize_index.test_database_schema", "database_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_index.test_database_schema", "schema_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_index.test_database_schema", "indexes.#", "0"),
					resource.TestCheckNoResourceAttr("data.materialize_index.test_all", "database_name"),
					resource.TestCheckNoResourceAttr("data.materialize_index.test_all", "schema_name"),
					// Cannot ensure the exact number of objects with parallel tests
					// Ensuring minimum
					resource.TestMatchResourceAttr("data.materialize_index.test_all", "indexes.#", regexp.MustCompile("([0-9]|\\d{2,})")),
				),
			},
		},
	})
}

func testAccDatasourceIndex(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_database" "test" {
		name    = "%[1]s"
	}

	resource "materialize_schema" "test" {
		name          = "%[1]s"
		database_name = materialize_database.test.name
	}

	data "materialize_index" "test_all" {}
	
	data "materialize_index" "test_database_schema" {
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
	}
	`, nameSpace)
}

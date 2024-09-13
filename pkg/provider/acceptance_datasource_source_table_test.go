package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSourceTable_basic(t *testing.T) {
	nameSpace := acctest.RandomWithPrefix("tf_test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSourceTable(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.materialize_source_table.test", "tables.0.name", fmt.Sprintf("%s_table", nameSpace)),
					resource.TestCheckResourceAttr("data.materialize_source_table.test", "tables.0.schema_name", "public"),
					resource.TestCheckResourceAttr("data.materialize_source_table.test", "tables.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("data.materialize_source_table.test", "tables.0.source.#", "1"),
					resource.TestCheckResourceAttr("data.materialize_source_table.test", "tables.0.source.0.name", fmt.Sprintf("%s_source", nameSpace)),
					resource.TestCheckResourceAttr("data.materialize_source_table.test", "tables.0.source.0.schema_name", "public"),
					resource.TestCheckResourceAttr("data.materialize_source_table.test", "tables.0.source.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("data.materialize_source_table.test", "tables.0.source_type", "load-generator"),
					resource.TestCheckResourceAttr("data.materialize_source_table.test", "tables.0.comment", "test comment"),
					resource.TestCheckResourceAttrSet("data.materialize_source_table.test", "tables.0.owner_name"),
				),
			},
		},
	})
}

func testAccDataSourceSourceTable(nameSpace string) string {
	return fmt.Sprintf(`
resource "materialize_source_load_generator" "test" {
	name                = "%[1]s_source"
	schema_name         = "public"
	database_name       = "materialize"
	load_generator_type = "AUCTION"
	auction_options {
		tick_interval = "1s"
	}
}

resource "materialize_source_table_load_generator" "test" {
	name           = "%[1]s_table"
	schema_name    = "public"
	database_name  = "materialize"
	source {
		name          = materialize_source_load_generator.test.name
		schema_name   = materialize_source_load_generator.test.schema_name
		database_name = materialize_source_load_generator.test.database_name
	}
	upstream_name = "bids"
	comment       = "test comment"
}

data "materialize_source_table" "test" {
	schema_name   = "public"
	database_name = "materialize"
	depends_on    = [materialize_source_table_load_generator.test]
}
`, nameSpace)
}

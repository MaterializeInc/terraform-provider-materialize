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
					resource.TestCheckResourceAttr("data.materialize_source_table.test", "tables.0.source_type", "postgres"),
					resource.TestCheckResourceAttr("data.materialize_source_table.test", "tables.0.comment", "test comment"),
					resource.TestCheckResourceAttrSet("data.materialize_source_table.test", "tables.0.owner_name"),
				),
			},
		},
	})
}

func testAccDataSourceSourceTable(nameSpace string) string {
	return fmt.Sprintf(`
resource "materialize_secret" "postgres_password" {
	name  = "%[1]s_secret_postgres"
	value = "c2VjcmV0Cg=="
}

resource "materialize_connection_postgres" "postgres_connection" {
	name    = "%[1]s_connection_postgres"
	host    = "postgres"
	port    = 5432
	user {
		text = "postgres"
	}
	password {
		name = materialize_secret.postgres_password.name
	}
	database = "postgres"
}

resource "materialize_source_postgres" "test" {
	name         = "%[1]s_source"
	cluster_name = "quickstart"

	postgres_connection {
		name = materialize_connection_postgres.postgres_connection.name
	}
	publication = "mz_source"
}

resource "materialize_source_table_postgres" "test" {
	name           = "%[1]s_table"
	schema_name    = "public"
	database_name  = "materialize"
	source {
		name          = materialize_source_postgres.test.name
		schema_name   = "public"
		database_name = "materialize"
	}
	upstream_name         = "table2"
	upstream_schema_name  = "public"
	comment               = "test comment"
}

data "materialize_source_table" "test" {
	schema_name   = "public"
	database_name = "materialize"
	depends_on    = [materialize_source_table_postgres.test]
}
`, nameSpace)
}

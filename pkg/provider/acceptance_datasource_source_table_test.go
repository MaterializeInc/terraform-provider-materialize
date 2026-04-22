package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceSourceTable_basic(t *testing.T) {
	nameSpace := acctest.RandomWithPrefix("tf_test")
	tableName := fmt.Sprintf("%s_table", nameSpace)
	sourceName := fmt.Sprintf("%s_source", nameSpace)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSourceTable(nameSpace),
				Check: checkSourceTableInDataSource(
					"data.materialize_source_table.test", tableName,
					map[string]string{
						"schema_name":            "public",
						"database_name":          "materialize",
						"source.#":               "1",
						"source.0.name":          sourceName,
						"source.0.schema_name":   "public",
						"source.0.database_name": "materialize",
						"source_type":            "postgres",
						"comment":                "test comment",
					},
					[]string{"owner_name"},
				),
			},
		},
	})
}

// checkSourceTableInDataSource locates the entry named `tableName` inside the
// `tables.*` list of a `materialize_source_table` data source and asserts the
// given attributes against it. The data source has no name filter and returns
// every source table in the schema, so indexing with `tables.0` is unreliable
// when parallel tests share `materialize.public`.
func checkSourceTableInDataSource(dataSource, tableName string, equals map[string]string, set []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSource]
		if !ok {
			return fmt.Errorf("data source %q not found in state", dataSource)
		}
		attrs := rs.Primary.Attributes
		count, err := strconv.Atoi(attrs["tables.#"])
		if err != nil {
			return fmt.Errorf("could not read tables.# on %s: %w", dataSource, err)
		}
		for i := 0; i < count; i++ {
			prefix := fmt.Sprintf("tables.%d.", i)
			if attrs[prefix+"name"] != tableName {
				continue
			}
			for k, want := range equals {
				if got := attrs[prefix+k]; got != want {
					return fmt.Errorf("%s%s = %q, want %q", prefix, k, got, want)
				}
			}
			for _, k := range set {
				if attrs[prefix+k] == "" {
					return fmt.Errorf("%s%s is empty, expected to be set", prefix, k)
				}
			}
			return nil
		}
		return fmt.Errorf("source table %q not found in %s (searched %d entries)", tableName, dataSource, count)
	}
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

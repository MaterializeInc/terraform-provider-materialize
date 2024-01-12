package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceSink_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceSink(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.materialize_sink.test_database_schema", "database_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_sink.test_database_schema", "schema_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_sink.test_database_schema", "sinks.#", "1"),
					// Cannot ensure the exact number of objects with parallel tests
					// Ensuring minimum
					resource.TestMatchResourceAttr("data.materialize_sink.test_all", "sinks.#", regexp.MustCompile("([1-9]|\\d{2,})")),
				),
			},
		},
	})
}

func testAccDatasourceSink(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_database" "test" {
		name    = "%[1]s"
	}

	resource "materialize_schema" "test" {
		name          = "%[1]s"
		database_name = materialize_database.test.name
	}

	resource "materialize_cluster" "test" {
		name = "%[1]s"
	}

	resource "materialize_connection_kafka" "test" {
		name = "%[1]s_conn"
		kafka_broker {
			broker = "redpanda:9092"
		}
		security_protocol = "PLAINTEXT"
	}

	resource "materialize_table" "test" {
		name = "%[1]s"
		column {
			name = "column_1"
			type = "text"
		}
		column {
			name = "column_2"
			type = "int"
		}
		column {
			name     = "column_3"
			type     = "text"
			nullable = true
		}
	}

	resource "materialize_sink_kafka" "a" {
		name          = "%[1]s_a"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name

		kafka_connection {
			name = materialize_connection_kafka.test.name
		}
		from {
			name = materialize_table.test.name
		}
		cluster_name = materialize_cluster.test.name
		topic = "sink_topic"
		format {
			json = true
		}
		envelope {
			debezium = true
		}
	}

	data "materialize_sink" "test_all" {
		depends_on    = [
			materialize_sink_kafka.a,
		]
	}

	data "materialize_sink" "test_database_schema" {
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		depends_on    = [
			materialize_sink_kafka.a,
		]
	}
	`, nameSpace)
}

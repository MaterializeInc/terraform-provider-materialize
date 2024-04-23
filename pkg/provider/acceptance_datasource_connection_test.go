package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceConnection_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceConnection(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.materialize_connection.test_database", "database_name", nameSpace),
					resource.TestCheckNoResourceAttr("data.materialize_connection.test_database", "schema_name"),
					resource.TestCheckResourceAttr("data.materialize_connection.test_database", "connections.#", "3"),
					resource.TestCheckResourceAttr("data.materialize_connection.test_database_schema", "database_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_connection.test_database_schema", "schema_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_connection.test_database_schema", "connections.#", "2"),
					resource.TestCheckResourceAttr("data.materialize_connection.test_database_2", "database_name", nameSpace+"_2"),
					resource.TestCheckNoResourceAttr("data.materialize_connection.test_database_2", "schema_name"),
					resource.TestCheckResourceAttr("data.materialize_connection.test_database_2", "connections.#", "2"),
					resource.TestCheckNoResourceAttr("data.materialize_connection.test_all", "database_name"),
					resource.TestCheckNoResourceAttr("data.materialize_connection.test_all", "schema_name"),
					// Cannot ensure the exact number of objects with parallel tests
					// Ensuring minimum
					resource.TestMatchResourceAttr("data.materialize_connection.test_all", "connections.#", regexp.MustCompile("([5-9]|\\d{2,})")),
				),
			},
		},
	})
}

func testAccDatasourceConnection(nameSpace string) string {
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

	resource "materialize_connection_kafka" "a" {
		name              = "%[1]s_a"
		database_name     = materialize_database.test.name
		schema_name       = materialize_schema.test.name
		security_protocol = "PLAINTEXT"
	  
		kafka_broker {
		  broker = "redpanda:9092"
		}
		validate = true
	}

	resource "materialize_connection_kafka" "b" {
		name              = "%[1]s_b"
		database_name     = materialize_database.test.name
		schema_name       = materialize_schema.test.name
		security_protocol = "PLAINTEXT"
	  
		kafka_broker {
		  broker = "redpanda:9092"
		}
		validate = true
	}

	resource "materialize_connection_kafka" "c" {
		name              = "%[1]s_c"
		database_name     = materialize_database.test.name
		schema_name       = materialize_schema.test.name
		security_protocol = "PLAINTEXT"
	  
		kafka_broker {
		  broker = "redpanda:9092"
		}
		validate = true
	}

	resource "materialize_connection_kafka" "d" {
		name              = "%[1]s_d"
		database_name     = materialize_database.test_2.name
		schema_name       = materialize_schema.public_schema2.name
		security_protocol = "PLAINTEXT"
	  
		kafka_broker {
		  broker = "redpanda:9092"
		}
		validate = true
	}

	resource "materialize_connection_kafka" "e" {
		name              = "%[1]s_e"
		database_name     = materialize_database.test_2.name
		schema_name       = materialize_schema.public_schema2.name
		security_protocol = "PLAINTEXT"
	  
		kafka_broker {
		  broker = "redpanda:9092"
		}
		validate = true
	}

	data "materialize_connection" "test_all" {
		depends_on    = [
			materialize_connection_kafka.a,
			materialize_connection_kafka.b,
			materialize_connection_kafka.c,
			materialize_connection_kafka.d,
			materialize_connection_kafka.e,
		]
	}

	data "materialize_connection" "test_database" {
		database_name = materialize_database.test.name
		depends_on    = [
			materialize_connection_kafka.a,
			materialize_connection_kafka.b,
			materialize_connection_kafka.c,
			materialize_connection_kafka.d,
			materialize_connection_kafka.e,
		]
	}
	
	data "materialize_connection" "test_database_schema" {
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		depends_on    = [
			materialize_connection_kafka.a,
			materialize_connection_kafka.b,
			materialize_connection_kafka.c,
			materialize_connection_kafka.d,
			materialize_connection_kafka.e,
		]
	}

	data "materialize_connection" "test_database_2" {
		database_name = materialize_database.test_2.name
		depends_on = [
			materialize_connection_kafka.a,
			materialize_connection_kafka.b,
			materialize_connection_kafka.c,
			materialize_connection_kafka.d,
			materialize_connection_kafka.e,
		]
	}
	`, nameSpace)
}

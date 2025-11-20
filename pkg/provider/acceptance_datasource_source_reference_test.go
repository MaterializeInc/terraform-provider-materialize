package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSourceReference_basic(t *testing.T) {
	addTestTopic()
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSourceReferenceConfig(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.materialize_source_reference.kafka", "source_id"),
					resource.TestCheckResourceAttrSet("data.materialize_source_reference.postgres", "source_id"),
					resource.TestCheckResourceAttrSet("data.materialize_source_reference.mysql", "source_id"),

					// Check total references
					resource.TestCheckResourceAttr("data.materialize_source_reference.kafka", "references.#", "1"),
					resource.TestCheckResourceAttr("data.materialize_source_reference.postgres", "references.#", "3"),
					resource.TestCheckResourceAttr("data.materialize_source_reference.mysql", "references.#", "4"),

					// Check Postgres reference attributes
					resource.TestCheckResourceAttr("data.materialize_source_reference.postgres", "references.0.namespace", "public"),
					resource.TestCheckResourceAttrSet("data.materialize_source_reference.postgres", "references.0.name"),
					resource.TestCheckResourceAttr("data.materialize_source_reference.postgres", "references.0.source_name", fmt.Sprintf("%s_source_postgres", nameSpace)),
					resource.TestCheckResourceAttr("data.materialize_source_reference.postgres", "references.0.source_type", "postgres"),
					resource.TestCheckResourceAttrSet("data.materialize_source_reference.postgres", "references.0.updated_at"),

					// Check MySQL reference attributes
					resource.TestCheckResourceAttr("data.materialize_source_reference.mysql", "references.0.namespace", "shop"),
					resource.TestCheckResourceAttrSet("data.materialize_source_reference.mysql", "references.0.name"),
					resource.TestCheckResourceAttr("data.materialize_source_reference.mysql", "references.0.source_name", fmt.Sprintf("%s_source_mysql", nameSpace)),
					resource.TestCheckResourceAttr("data.materialize_source_reference.mysql", "references.1.source_type", "mysql"),
					resource.TestCheckResourceAttrSet("data.materialize_source_reference.mysql", "references.1.updated_at"),

					// Check Kafka reference attributes
					resource.TestCheckResourceAttr("data.materialize_source_reference.kafka", "references.0.name", "terraform"),
					resource.TestCheckResourceAttr("data.materialize_source_reference.kafka", "references.0.source_name", fmt.Sprintf("%s_source_kafka", nameSpace)),
					resource.TestCheckResourceAttr("data.materialize_source_reference.kafka", "references.0.source_type", "kafka"),
					resource.TestCheckResourceAttrSet("data.materialize_source_reference.kafka", "references.0.updated_at"),
				),
			},
		},
	})
}

func testAccDataSourceSourceReferenceConfig(nameSpace string) string {
	return fmt.Sprintf(`
	// Postgres setup
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

	resource "materialize_source_postgres" "test_source_postgres" {
		name         = "%[1]s_source_postgres"
		cluster_name = "quickstart"

		postgres_connection {
			name = materialize_connection_postgres.postgres_connection.name
		}
		publication = "mz_source"
		table {
			upstream_name  = "table1"
			upstream_schema_name = "public"
			name = "%[1]s_table1"
		}
		table {
			upstream_name  = "table2"
			upstream_schema_name = "public"
			name = "%[1]s_table2"
		}
		table {
			upstream_name  = "table3"
			upstream_schema_name = "public"
			name = "%[1]s_table3"
		}
	}

	// MySQL setup
	resource "materialize_secret" "mysql_password" {
		name  = "%[1]s_secret_mysql"
		value = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_mysql" "mysql_connection" {
		name    = "%[1]s_connection_mysql"
		host    = "mysql"
		port    = 3306
		user {
			text = "repluser"
		}
		password {
			name = materialize_secret.mysql_password.name
		}
	}

	resource "materialize_source_mysql" "test_source_mysql" {
		name         = "%[1]s_source_mysql"
		cluster_name = "quickstart"

		mysql_connection {
			name = materialize_connection_mysql.mysql_connection.name
		}
		text_columns = ["shop.mysql_table4.status"]
		table {
			upstream_name        = "mysql_table1"
			upstream_schema_name = "shop"
		}
		table {
			upstream_name        = "mysql_table2"
			upstream_schema_name = "shop"
		}
		table {
			upstream_name        = "mysql_table3"
			upstream_schema_name = "shop"
		}
		table {
			upstream_name        = "mysql_table4"
			upstream_schema_name = "shop"
		}
	}

	// Kafka setup
	resource "materialize_connection_kafka" "kafka_connection" {
		name = "%[1]s_connection_kafka"
		kafka_broker {
			broker = "redpanda:9092"
		}
		security_protocol = "PLAINTEXT"
	}

	resource "materialize_source_kafka" "test_source_kafka" {
		name         = "%[1]s_source_kafka"
		cluster_name = "quickstart"
		topic        = "terraform"

		kafka_connection {
			name = materialize_connection_kafka.kafka_connection.name
		}
		value_format {
			json = true
		}
		key_format {
			json = true
		}
	}

	data "materialize_source_reference" "kafka" {
		source_id = materialize_source_kafka.test_source_kafka.id
		depends_on = [
			materialize_source_kafka.test_source_kafka
		]
	}

	data "materialize_source_reference" "postgres" {
		source_id = materialize_source_postgres.test_source_postgres.id
		depends_on = [
			materialize_source_postgres.test_source_postgres
		]
	}

	data "materialize_source_reference" "mysql" {
		source_id = materialize_source_mysql.test_source_mysql.id
		depends_on = [
			materialize_source_mysql.test_source_mysql
		]
	}
	`, nameSpace)
}

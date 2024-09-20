package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSourceTableKafka_basic(t *testing.T) {
	addTestTopic()
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableKafkaBasicResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table_kafka.test_kafka"),
					resource.TestMatchResourceAttr("materialize_source_table_kafka.test_kafka", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "name", nameSpace+"_table_kafka"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s_table_kafka"`, nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "upstream_name", "terraform"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "include_key", "true"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "include_key_alias", "message_key"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "include_headers", "true"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "include_headers_alias", "message_headers"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "include_partition", "true"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "include_partition_alias", "message_partition"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "include_offset", "true"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "include_offset_alias", "message_offset"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "include_timestamp", "true"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "include_timestamp_alias", "message_timestamp"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "format.0.json", "true"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "key_format.0.text", "true"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "value_format.0.json", "true"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "envelope.0.upsert", "true"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "envelope.0.upsert_options.0.value_decoding_errors.0.inline.0.enabled", "true"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "envelope.0.upsert_options.0.value_decoding_errors.0.inline.0.alias", "decoding_error"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "expose_progress.0.name", nameSpace+"_progress"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "comment", "This is a test Kafka source table"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "source.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "source.0.name", nameSpace+"_source_kafka"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "source.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test_kafka", "source.0.database_name", "materialize"),
				),
			},
		},
	})
}

func TestAccSourceTableKafka_update(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableKafkaResource(nameSpace, "kafka_table2", "mz_system", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table_kafka.test"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test", "name", nameSpace+"_table"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test", "upstream_name", "kafka_table2"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test", "comment", ""),
				),
			},
			{
				Config: testAccSourceTableKafkaResource(nameSpace, "terraform", nameSpace+"_role", "Updated comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table_kafka.test"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test", "name", nameSpace+"_table"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test", "upstream_name", "terraform"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test", "ownership_role", nameSpace+"_role"),
					resource.TestCheckResourceAttr("materialize_source_table_kafka.test", "comment", "Updated comment"),
				),
			},
		},
	})
}

func TestAccSourceTableKafka_disappears(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceTableDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableKafkaResource(nameSpace, "kafka_table2", "mz_system", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table_kafka.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "TABLE",
							Name:       nameSpace + "_table",
						},
					),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSourceTableKafkaBasicResource(nameSpace string) string {
	return fmt.Sprintf(`
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
	}

	resource "materialize_source_table_kafka" "test_kafka" {
		name           = "%[1]s_table_kafka"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name = materialize_source_kafka.test_source_kafka.name
		}

		upstream_name = "terraform"
		include_key   = true
		include_key_alias = "message_key"
		include_headers = true
		include_headers_alias = "message_headers"
		include_partition = true
		include_partition_alias = "message_partition"
		include_offset = true
		include_offset_alias = "message_offset"
		include_timestamp = true
		include_timestamp_alias = "message_timestamp"


		key_format {
			text = true
		}
		value_format {
			json = true
		}

		envelope {
			upsert = true
			upsert_options {
				value_decoding_errors {
					inline {
						enabled = true
						alias = "decoding_error"
					}
				}
			}
		}

		ownership_role = "mz_system"
		comment = "This is a test Kafka source table"
	}
	`, nameSpace)
}

func testAccSourceTableKafkaResource(nameSpace, upstreamName, ownershipRole, comment string) string {
	return fmt.Sprintf(`
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

		key_format {
			json = true
		}
		value_format {
			json = true
		}
	}

	resource "materialize_role" "test_role" {
		name = "%[1]s_role"
	}

	resource "materialize_source_table_kafka" "test" {
		name           = "%[1]s_table"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name          = materialize_source_kafka.test_source_kafka.name
			schema_name   = "public"
			database_name = "materialize"
		}

		upstream_name = "%[2]s"

		ownership_role = "%[3]s"
		comment        = "%[4]s"

		depends_on = [materialize_role.test_role]
	}
	`, nameSpace, upstreamName, ownershipRole, comment)
}

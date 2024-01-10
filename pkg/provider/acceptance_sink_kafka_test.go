package provider

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSinkKafka_basic(t *testing.T) {
	sinkName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	sink2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	tableName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSinkKafkaResource(roleName, connName, tableName, sinkName, sink2Name, roleName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkKafkaExists("materialize_sink_kafka.test"),
					resource.TestMatchResourceAttr("materialize_sink_kafka.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "name", sinkName),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, sinkName)),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "topic", "sink_topic"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "envelope.0.debezium", "true"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.json", "true"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "ownership_role", "mz_system"),
					testAccCheckSinkKafkaExists("materialize_sink_kafka.test_role"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test_role", "name", sink2Name),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test_role", "comment", "Comment"),
				),
			},
			{
				ResourceName:      "materialize_sink_kafka.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSinkKafkaAvro_basic(t *testing.T) {
	sinkName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSinkKafkaAvroResource(sinkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkKafkaExists("materialize_sink_kafka.test"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "name", sinkName+"_sink"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, sinkName+"_sink")),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "cluster_name", sinkName+"_cluster"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "topic", "topic1"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "key.0", "counter"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "key_not_enforced", "true"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.schema_registry_connection.0.name", sinkName+"_conn_schema"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.schema_registry_connection.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.schema_registry_connection.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_type.0.object.0.name", sinkName+"_load_gen"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_type.0.object.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_type.0.object.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_type.0.doc", "top level comment"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_column.0.object.0.name", sinkName+"_load_gen"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_column.0.object.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_column.0.object.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_column.0.column", "counter"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_column.0.doc", "comment key"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_column.0.key", "true"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_column.1.object.0.name", sinkName+"_load_gen"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_column.1.object.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_column.1.object.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_column.1.column", "counter"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_column.1.doc", "comment value"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.avro.0.avro_doc_column.1.value", "true"),
				),
			},
			{
				ResourceName:      "materialize_sink_kafka.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSinkKafka_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	sinkName := fmt.Sprintf("old_%s", slug)
	newSinkName := fmt.Sprintf("new_%s", slug)
	sink2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	tableName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSinkKafkaResource(roleName, connName, tableName, sinkName, sink2Name, "mz_system", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkKafkaExists("materialize_sink_kafka.test"),
					testAccCheckSinkKafkaExists("materialize_sink_kafka.test_role"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test_role", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test_role", "comment", "Comment"),
				),
			},
			{
				Config: testAccSinkKafkaResource(roleName, connName, tableName, newSinkName, sink2Name, roleName, "New Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkKafkaExists("materialize_sink_kafka.test"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "name", newSinkName),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newSinkName)),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "topic", "sink_topic"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "envelope.0.debezium", "true"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.json", "true"),
					testAccCheckSinkKafkaExists("materialize_sink_kafka.test_role"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test_role", "comment", "New Comment"),
				),
			},
		},
	})
}

func TestAccSinkKafka_disappears(t *testing.T) {
	sinkName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	sink2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	tableName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSinkKafkaDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSinkKafkaResource(roleName, connName, tableName, sinkName, sink2Name, roleName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkKafkaExists("materialize_sink_kafka.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "SINK",
							Name:       sinkName,
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSinkKafkaResource(roleName, connName, tableName, sinkName, sink2Name, sinkOwner, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "test" {
		name = "%[1]s"
	}

	resource "materialize_connection_kafka" "test" {
		name = "%[2]s"
		kafka_broker {
			broker = "redpanda:9092"
		}
		security_protocol = "PLAINTEXT"
	}

	resource "materialize_table" "test" {
		name = "%[3]s"
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

	resource "materialize_sink_kafka" "test" {
		name = "%[4]s"
		kafka_connection {
			name = materialize_connection_kafka.test.name
		}
		from {
			name = materialize_table.test.name
		}
		size  = "3xsmall"
		topic = "sink_topic"
		compression_type = "none"
		format {
			json = true
		}
		envelope {
			debezium = true
		}
	}

	resource "materialize_sink_kafka" "test_role" {
		name = "%[5]s"
		kafka_connection {
			name = materialize_connection_kafka.test.name
		}
		from {
			name = materialize_table.test.name
		}
		size  = "3xsmall"
		topic = "sink_topic"
		format {
			json = true
		}
		envelope {
			debezium = true
		}

		ownership_role = "%[6]s"
		comment = "%[7]s"

		depends_on = [materialize_role.test]
	}
	`, roleName, connName, tableName, sinkName, sink2Name, sinkOwner, comment)
}

func testAccSinkKafkaAvroResource(sinkName string) string {
	return fmt.Sprintf(`
	resource "materialize_cluster" "test" {
		name = "%[1]s_cluster"
		size = "3xsmall"
	}

	resource "materialize_source_load_generator" "test" {
		name                = "%[1]s_load_gen"
		size                = "3xsmall"
		load_generator_type = "COUNTER"
	}

	resource "materialize_connection_kafka" "test" {
		name              = "%[1]s_conn"
		security_protocol = "PLAINTEXT"
		kafka_broker {
			broker = "redpanda:9092"
		}
		validate = true
	}

	resource "materialize_connection_confluent_schema_registry" "test" {
		name    = "%[1]s_conn_schema"
		url     = "http://redpanda:8081"
	}

	resource "materialize_sink_kafka" "test" {
		name             = "%[1]s_sink"
		cluster_name     = materialize_cluster.test.name
		topic            = "topic1"
		compression_type = "none"
		key              = ["counter"]
		key_not_enforced = true
		from {
		  name          = materialize_source_load_generator.test.name
		  database_name = materialize_source_load_generator.test.database_name
		  schema_name   = materialize_source_load_generator.test.schema_name
		}
		kafka_connection {
		  name          = materialize_connection_kafka.test.name
		  database_name = materialize_connection_kafka.test.database_name
		  schema_name   = materialize_connection_kafka.test.schema_name
		}
		format {
			avro {
				schema_registry_connection {
					name          = materialize_connection_confluent_schema_registry.test.name
					database_name = materialize_connection_confluent_schema_registry.test.database_name
					schema_name   = materialize_connection_confluent_schema_registry.test.schema_name
				}
				avro_doc_type {
					object {
						name          = materialize_source_load_generator.test.name
						database_name = materialize_source_load_generator.test.database_name
						schema_name   = materialize_source_load_generator.test.schema_name
					}
					doc = "top level comment"
				}
				avro_doc_column {
					object {
						name          = materialize_source_load_generator.test.name
						database_name = materialize_source_load_generator.test.database_name
						schema_name   = materialize_source_load_generator.test.schema_name
					}
					column = "counter"
					doc    = "comment key"
					key    = true
				}
				avro_doc_column {
					object {
						name          = materialize_source_load_generator.test.name
						database_name = materialize_source_load_generator.test.database_name
						schema_name   = materialize_source_load_generator.test.schema_name
					}
					column = "counter"
					doc    = "comment value"
					value  = true
				}
			}
		}
		envelope {
			debezium = true
		}
	  }
	`, sinkName)
}

func testAccCheckSinkKafkaExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("sink kafka not found: %s", name)
		}
		_, err = materialize.ScanSink(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckSinkKafkaDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		_, err = db.Exec(fmt.Sprintf(`DROP SINK "%s";`, name))
		return err
	}
}

func testAccCheckAllSinkKafkaDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_sink_kafka" {
			continue
		}

		_, err := materialize.ScanSink(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("sink %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

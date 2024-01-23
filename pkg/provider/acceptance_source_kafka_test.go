package provider

import (
	"database/sql"
	"fmt"
	"os/exec"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Initialize a topic used by Kafka Testacc against the docker compose
func addTestTopic() error {
	cmd := exec.Command("docker", "exec", "redpanda", "rpk", "topic", "create", "terraform")
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func TestAccSourceKafka_basic(t *testing.T) {
	addTestTopic()
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	source2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceKafkaResource(roleName, connName, sourceName, source2Name, roleName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceKafkaExists("materialize_source_kafka.test"),
					resource.TestMatchResourceAttr("materialize_source_kafka.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, sourceName)),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "topic", "terraform"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "key_format.0.text", "true"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "value_format.0.text", "true"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "envelope.0.none", "true"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "envelope.0.debezium", "false"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "envelope.0.upsert", "false"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "kafka_connection.0.name", connName),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "kafka_connection.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "kafka_connection.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "start_offset.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "include_timestamp_alias", "timestamp_alias"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "include_offset", "true"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "include_offset_alias", "offset_alias"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "include_partition", "true"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "include_partition_alias", "partition_alias"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "include_key_alias", "key_alias"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "subsource.#", "1"),
					testAccCheckSourceKafkaExists("materialize_source_kafka.test_role"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test_role", "name", source2Name),
					resource.TestCheckResourceAttr("materialize_source_kafka.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_source_kafka.test_role", "comment", "Comment"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test_role", "subsource.#", "1"),
				),
			},
			{
				ResourceName:      "materialize_source_kafka.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSourceKafkaAvro_basic(t *testing.T) {
	addTestTopic()
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceKafkaResourceAvro(sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceKafkaExists("materialize_source_kafka.test"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "name", sourceName+"_source"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, sourceName+"_source")),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "cluster_name", sourceName+"_cluster"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "topic", "terraform"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "envelope.0.none", "true"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "kafka_connection.0.name", sourceName+"_conn"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "kafka_connection.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "kafka_connection.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "format.0.avro.0.schema_registry_connection.0.name", sourceName+"_conn_schema"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "format.0.avro.0.schema_registry_connection.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "format.0.avro.0.schema_registry_connection.0.schema_name", "public"),
				),
			},
			{
				ResourceName:      "materialize_source_kafka.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSourceKafka_update(t *testing.T) {
	addTestTopic()
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	sourceName := fmt.Sprintf("old_%s", slug)
	newSourceName := fmt.Sprintf("new_%s", slug)
	source2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceKafkaResource(roleName, connName, sourceName, source2Name, "mz_system", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceKafkaExists("materialize_source_kafka.test"),
					testAccCheckSourceKafkaExists("materialize_source_kafka.test_role"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, sourceName)),
					resource.TestCheckResourceAttr("materialize_source_kafka.test_role", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test_role", "comment", "Comment"),
				),
			},
			{
				Config: testAccSourceKafkaResource(roleName, connName, newSourceName, source2Name, roleName, "New Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceKafkaExists("materialize_source_kafka.test"),
					testAccCheckSourceKafkaExists("materialize_source_kafka.test_role"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "name", newSourceName),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newSourceName)),
					resource.TestCheckResourceAttr("materialize_source_kafka.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_source_kafka.test_role", "comment", "New Comment"),
				),
			},
		},
	})
}

func TestAccSourceKafka_disappears(t *testing.T) {
	addTestTopic()
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	source2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceKafkaDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceKafkaResource(roleName, connName, sourceName, source2Name, roleName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceKafkaExists("materialize_source_kafka.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "SOURCE",
							Name:       sourceName,
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSourceKafkaResource(roleName, connName, sourceName, source2Name, sourceOwner, comment string) string {
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

	resource "materialize_source_kafka" "test" {
		name = "%[3]s"
		kafka_connection {
			name = materialize_connection_kafka.test.name
		}

		cluster_name = "quickstart"
		topic = "terraform"
		key_format {
			text = true
		}
		value_format {
			text = true
		}
		envelope {
			none = true
		}

		start_offset = [0]
		include_timestamp_alias = "timestamp_alias"
		include_offset = true
		include_offset_alias = "offset_alias"
		include_partition = true
		include_partition_alias = "partition_alias"
		include_key_alias = "key_alias"
	}

	resource "materialize_source_kafka" "test_role" {
		name = "%[4]s"
		cluster_name = "quickstart"
		topic = "terraform"

		kafka_connection {
			name = materialize_connection_kafka.test.name
		}
		key_format {
			text = true
		}
		value_format {
			text = true
		}
		envelope {
			none = true
		}
		ownership_role = "%[5]s"
		comment = "%[6]s"

		depends_on = [materialize_role.test]
	}
`, roleName, connName, sourceName, source2Name, sourceOwner, comment)
}

func testAccSourceKafkaResourceAvro(sourceName string) string {
	return fmt.Sprintf(`
	resource "materialize_connection_kafka" "test" {
		name = "%[1]s_conn"
		kafka_broker {
			broker = "redpanda:9092"
		}
		security_protocol = "PLAINTEXT"
	}

	resource "materialize_connection_confluent_schema_registry" "test" {
		name = "%[1]s_conn_schema"
		url  = "http://redpanda:8081"
	}

	resource "materialize_cluster" "test" {
		name               = "%[1]s_cluster"
		size               = "3xsmall"
	}

	resource "materialize_source_load_generator" "test" {
		name                = "%[1]s_load_gen"
		cluster_name        = materialize_cluster.test.name
		load_generator_type = "COUNTER"
	}

	resource "materialize_sink_kafka" "test" {
		name             = "%[1]s_sink"
		topic            = "terraform"
		cluster_name	 = materialize_cluster.test.name
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
			}
		}
		envelope {
			debezium = true
		}
	}

	resource "materialize_source_kafka" "test" {
		name = "%[1]s_source"
		cluster_name = materialize_cluster.test.name
		topic = "terraform"
		kafka_connection {
			name          = materialize_connection_kafka.test.name
			schema_name   = materialize_connection_kafka.test.schema_name
			database_name = materialize_connection_kafka.test.database_name
		}
		format {
			avro {
				schema_registry_connection {
					name          = materialize_connection_confluent_schema_registry.test.name
					schema_name   = materialize_connection_confluent_schema_registry.test.schema_name
					database_name = materialize_connection_confluent_schema_registry.test.database_name
				}
			}
		}
		envelope {
			none = true
		}
		start_timestamp = -1000
		depends_on = [materialize_sink_kafka.test]
	}
`, sourceName)
}

func testAccCheckSourceKafkaExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("source kafka not found: %s", name)
		}
		_, err = materialize.ScanSource(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllSourceKafkaDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_source_kafka" {
			continue
		}

		_, err := materialize.ScanSource(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("source %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

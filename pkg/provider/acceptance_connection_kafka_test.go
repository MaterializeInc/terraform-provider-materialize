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

func TestAccConnKafka_basic(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaResource(roleName, connectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.test"),
					resource.TestMatchResourceAttr("materialize_connection_kafka.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.#", "1"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.0.broker", "redpanda:9092"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "comment", "object comment"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "progress_topic", "progress_topic"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "progress_topic_replication_factor", "1"),
					testAccCheckConnKafkaExists("materialize_connection_kafka.test_role"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test_role", "name", connection2Name),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test_role", "ownership_role", roleName),
				),
			},
			{
				ResourceName:      "materialize_connection_kafka.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccConnKafkaMultipleBrokers_basic(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaMultipleBrokerResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.test"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.#", "3"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.0.broker", "redpanda:9092"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.1.broker", "redpanda:9092"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.2.broker", "redpanda:9092"),
				),
			},
			{
				ResourceName:      "materialize_connection_kafka.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccConnKafkaMultipleSsh_basic(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaSshResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.test"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.#", "3"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.0.broker", "redpanda:9092"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.0.ssh_tunnel.0.name", connectionName+"_ssh_conn_1"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.1.broker", "redpanda:9092"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.1.ssh_tunnel.0.name", connectionName+"_ssh_conn_2"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.2.broker", "redpanda:9092"),
					resource.TestCheckNoResourceAttr("materialize_connection_kafka.test", "kafka_broker.2.ssh_tunnel.0.name"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "ssh_tunnel.0.name", connectionName+"_ssh_conn_2"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "ssh_tunnel.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "ssh_tunnel.0.schema_name", "public"),
				),
			},
			{
				ResourceName:      "materialize_connection_kafka.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccConnKafkaAwsPrivatelink_basic(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaAwsPrivatelinkResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.privatelink_top_level"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.privatelink_top_level", "name", connectionName+"_top_level_privatelink"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.privatelink_top_level", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.privatelink_top_level", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.privatelink_top_level", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName+"_top_level_privatelink")),
					resource.TestCheckResourceAttr("materialize_connection_kafka.privatelink_top_level", "aws_privatelink.#", "1"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_privatelink", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_privatelink", "kafka_broker.#", "1"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_privatelink", "kafka_broker.0.broker", "b-1.hostname-1:9096"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_privatelink", "kafka_broker.0.target_group_port", "9001"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_privatelink", "kafka_broker.0.privatelink_connection.#", "1"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_privatelink", "kafka_broker.0.privatelink_connection.0.name", "privatelink_conn"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_privatelink", "kafka_broker.0.privatelink_connection.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_privatelink", "kafka_broker.0.privatelink_connection.0.schema_name", "public"),
				),
			},
			{
				ResourceName:      "materialize_connection_kafka.privatelink_top_level",
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				ResourceName:      "materialize_connection_kafka.kafka_privatelink",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccConnKafka_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	connectionName := fmt.Sprintf("old_%s", slug)
	newConnectionName := fmt.Sprintf("new_%s", slug)
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaResource(roleName, connectionName, connection2Name, "mz_system"),
			},
			{
				Config: testAccConnKafkaResource(roleName, newConnectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.test"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "name", newConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newConnectionName)),
					testAccCheckConnKafkaExists("materialize_connection_kafka.test_role"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test_role", "ownership_role", roleName),
				),
			},
		},
	})
}

func TestAccConnKafka_updateExtended(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaSASLResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.kafka_connection"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_connection", "security_protocol", "SASL_SSL"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_connection", "sasl_mechanisms", "SCRAM-SHA-512"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_connection", "kafka_broker.#", "2"),
				),
			},
			{
				Config: testAccConnKafkaSSLResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.kafka_connection2"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_connection2", "security_protocol", "SSL"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_connection2", "kafka_broker.#", "2"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_connection2", "ssl_certificate.0.text", "certificate-content"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_connection2", "ssl_certificate_authority.0.text", "ca-content"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_connection2", "ssl_key.0.name", connectionName+"_ssl_secret"),
				),
			},
		},
	})
}

func TestAccConnKafka_updateBrokersToPrivatelink(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaBrokerResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.kafka_broker_conn"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_broker_conn", "kafka_broker.#", "2"),
					resource.TestCheckNoResourceAttr("materialize_connection_kafka.kafka_broker_conn", "aws_privatelink.#"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_broker_conn", "kafka_broker.0.broker", "redpanda:9092"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_broker_conn", "kafka_broker.1.broker", "redpanda:9093"),
				),
			},
			{
				Config: testAccConnKafkaPrivatelinkResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.kafka_broker_conn"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.kafka_broker_conn", "aws_privatelink.#", "1"),
					resource.TestCheckNoResourceAttr("materialize_connection_kafka.kafka_broker_conn", "kafka_broker.#"),
				),
			},
		},
	})
}

func TestAccConnKafkaAwsIAM_basic(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	awsConnectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resourceName := "materialize_connection_kafka.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllConnKafkaDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaAwsIAMResource(connectionName, awsConnectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "database_name", "materialize"),
					resource.TestCheckResourceAttr(resourceName, "schema_name", "public"),
					resource.TestCheckResourceAttr(resourceName, "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
					resource.TestCheckResourceAttr(resourceName, "kafka_broker.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_broker.0.broker", "msk.mycorp.com:9092"),
					resource.TestCheckResourceAttr(resourceName, "security_protocol", "SASL_SSL"),
					resource.TestCheckResourceAttr(resourceName, "aws_connection.0.name", awsConnectionName),
					resource.TestCheckResourceAttr(resourceName, "aws_connection.0.database_name", "materialize"),
					resource.TestCheckResourceAttr(resourceName, "aws_connection.0.schema_name", "public"),
				),
			},
		},
	})
}

func testAccConnKafkaBrokerResource(connectionName string) string {
	return fmt.Sprintf(`
resource "materialize_connection_kafka" "kafka_broker_conn" {
  name = "%s"
  kafka_broker {
    broker = "redpanda:9092"
  }
  kafka_broker {
    broker = "redpanda:9093"
  }
  validate = false
}
`, connectionName)
}

func testAccConnKafkaPrivatelinkResource(connectionName string) string {
	return fmt.Sprintf(`
resource "materialize_connection_kafka" "kafka_broker_conn" {
  name = "%s"
  aws_privatelink {
    privatelink_connection {
      name = "privatelink_conn"
      database_name = "materialize"
      schema_name = "public"
    }
    privatelink_connection_port = 9092
  }
  validate = false
}
`, connectionName)
}

func testAccConnKafkaSASLResource(connectionName string) string {
	return fmt.Sprintf(`
resource "materialize_secret" "sasl_password" {
	name  = "%[1]s_sasl_password"
	value = "sasl_password"
}

resource "materialize_connection_kafka" "kafka_connection" {
  name              = "%[1]s"
  security_protocol = "SASL_SSL"
  sasl_mechanisms   = "SCRAM-SHA-512"
  sasl_username {
   text = "sasl_username"
  }
  sasl_password {
    name          = materialize_secret.sasl_password.name
    database_name = materialize_secret.sasl_password.database_name
    schema_name   = materialize_secret.sasl_password.schema_name
  }
  kafka_broker {
    broker = "redpanda:9092"
  }
  kafka_broker {
    broker = "redpanda:9093"
  }
  validate = false
}
`, connectionName)
}

func testAccConnKafkaSSLResource(connectionName string) string {
	return fmt.Sprintf(`
resource "materialize_secret" "ssl_secret" {
	name  = "%[1]s_ssl_secret"
	value = "ssl_secret"
}

resource "materialize_connection_ssh_tunnel" "ssh_connection_extended" {
	name = "%[1]s_ssh_conn_extended"
	host = "ssh_host"
	user = "ssh_user"
	port = 22

	validate = false
}

resource "materialize_connection_kafka" "kafka_connection2" {
  name              = "%[1]s"
  comment           = "connection kafka comment"
  security_protocol = "SSL"
  ssl_certificate {
    text = "certificate-content"
  }
  ssl_key {
    name = materialize_secret.ssl_secret.name
	database_name = materialize_secret.ssl_secret.database_name
	schema_name = materialize_secret.ssl_secret.schema_name
  }
  ssl_certificate_authority {
    text = "ca-content"
  }
  kafka_broker {
    broker = "redpanda:9093"
  }
  kafka_broker {
    broker = "redpanda:9092"
    ssh_tunnel {
      name          = materialize_connection_ssh_tunnel.ssh_connection_extended.name
      database_name = materialize_connection_ssh_tunnel.ssh_connection_extended.database_name
      schema_name   = materialize_connection_ssh_tunnel.ssh_connection_extended.schema_name
    }
  }
  validate = false
}
`, connectionName)
}

func testAccConnKafkaAwsIAMResource(connectionName, awsConnectionName string) string {
	return fmt.Sprintf(`
resource "materialize_connection_aws" "test" {
    name                    = "%[2]s"
    assume_role_arn         = "arn:aws:iam::400121260767:role/MaterializeMSK"
    assume_role_session_name = "materialize-session"
}

resource "materialize_connection_kafka" "test" {
    name = "%[1]s"
    kafka_broker {
        broker = "msk.mycorp.com:9092"
    }
    security_protocol = "SASL_SSL"
    aws_connection {
        name          = materialize_connection_aws.test.name
        database_name = materialize_connection_aws.test.database_name
        schema_name   = materialize_connection_aws.test.schema_name
    }
	validate = false
}
`, connectionName, awsConnectionName)
}

func TestAccConnKafka_disappears(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllConnKafkaDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaResource(roleName, connectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: materialize.BaseConnection,
							Name:       connectionName,
						},
					),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConnKafkaResource(roleName, connectionName, connection2Name, connectionOwner string) string {
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
    comment = "object comment"
    progress_topic = "progress_topic"
    progress_topic_replication_factor = 1
}

resource "materialize_connection_kafka" "test_role" {
    name = "%[3]s"
    kafka_broker {
        broker = "redpanda:9092"
    }
    security_protocol = "PLAINTEXT"
    ownership_role = "%[4]s"

    depends_on = [materialize_role.test]

    validate = false
}
`, roleName, connectionName, connection2Name, connectionOwner)
}

func testAccConnKafkaMultipleBrokerResource(connectionName string) string {
	return fmt.Sprintf(`
	resource "materialize_connection_kafka" "test" {
		name = "%[1]s"
		kafka_broker {
			broker = "redpanda:9092"
		}
		kafka_broker {
			broker = "redpanda:9092"
		}
		kafka_broker {
			broker = "redpanda:9092"
		}
		security_protocol = "PLAINTEXT"
		validate = false
	}
	`, connectionName)
}

func testAccConnKafkaSshResource(connectionName string) string {
	return fmt.Sprintf(`
	resource "materialize_connection_ssh_tunnel" "ssh_connection_1" {
		name = "%[1]s_ssh_conn_1"
		host = "ssh_host"
		user = "ssh_user"
		port = 22

		validate = false
	}

	resource "materialize_connection_ssh_tunnel" "ssh_connection_2" {
		name = "%[1]s_ssh_conn_2"
		host = "ssh_host"
		user = "ssh_user"
		port = 22

		validate = false
	}

	resource "materialize_connection_kafka" "test" {
		name = "%[1]s"
		kafka_broker {
			broker = "redpanda:9092"
			ssh_tunnel {
				name = materialize_connection_ssh_tunnel.ssh_connection_1.name
			}
		}
		kafka_broker {
			broker = "redpanda:9092"
			ssh_tunnel {
				name = materialize_connection_ssh_tunnel.ssh_connection_2.name
			}
		}
		kafka_broker {
			broker = "redpanda:9092"
		}
		ssh_tunnel {
			name = materialize_connection_ssh_tunnel.ssh_connection_2.name
		}
		security_protocol = "PLAINTEXT"
		validate = false
	}
	`, connectionName)
}

// Top level privatelink
func testAccConnKafkaAwsPrivatelinkResource(connectionName string) string {
	return fmt.Sprintf(`
	resource "materialize_connection_kafka" "privatelink_top_level" {
		name = "%[1]s_top_level_privatelink"
		aws_privatelink {
			privatelink_connection {
				name          = "privatelink_conn"
			  	database_name = "materialize"
			  	schema_name   = "public"
			}
			privatelink_connection_port = 9092
		}
		validate = false
	}

	resource "materialize_connection_kafka" "kafka_privatelink" {
		name = "%[1]s"
		kafka_broker {
			broker            = "b-1.hostname-1:9096"
			target_group_port = "9001"
			availability_zone = "use1-az2"
			privatelink_connection {
			  	name          = "privatelink_conn"
			  	database_name = "materialize"
			  	schema_name   = "public"
			}
		}
		validate = false
	}
	`, connectionName)
}

func testAccCheckConnKafkaExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("connection kafka not found: %s", name)
		}
		_, err = materialize.ScanConnection(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllConnKafkaDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_connection_kafka" {
			continue
		}

		_, err := materialize.ScanConnection(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("connection %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

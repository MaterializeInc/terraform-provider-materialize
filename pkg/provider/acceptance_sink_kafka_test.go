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
	"github.com/jmoiron/sqlx"
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

func testAccCheckSinkKafkaExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("sink kafka not found: %s", name)
		}
		_, err := materialize.ScanSink(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckSinkKafkaDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP SINK "%s";`, name))
		return err
	}
}

func testAccCheckAllSinkKafkaDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

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

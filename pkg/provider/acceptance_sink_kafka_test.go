package provider

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jmoiron/sqlx"
)

func TestAccSinkKafka_basic(t *testing.T) {
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	sinkName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	tableName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSinkKafkaResource(connName, sinkName, tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkKafkaExists("materialize_sink_kafka.test"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "name", sinkName),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, sinkName)),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "topic", "sink_topic"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "envelope.0.debezium", "true"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.json", "true"),
				),
			},
		},
	})
}

func TestAccSinkKafka_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	connName := fmt.Sprintf("conn_%s", slug)
	tableName := fmt.Sprintf("table_%s", slug)
	sinkName := fmt.Sprintf("old_%s", slug)
	newSinkName := fmt.Sprintf("new_%s", slug)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSinkKafkaResource(connName, sinkName, tableName),
			},
			{
				Config: testAccSinkKafkaResource(connName, newSinkName, tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkKafkaExists("materialize_sink_kafka.test"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "name", newSinkName),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newSinkName)),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "topic", "sink_topic"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "envelope.0.debezium", "true"),
					resource.TestCheckResourceAttr("materialize_sink_kafka.test", "format.0.json", "true"),
				),
			},
		},
	})
}

func TestAccSinkKafka_disappears(t *testing.T) {
	sinkName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	tableName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSinkKafkaDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSinkKafkaResource(connName, sinkName, tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkKafkaExists("materialize_sink_kafka.test"),
					testAccCheckSinkKafkaDisappears(sinkName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSinkKafkaResource(connName string, sinkName string, tableName string) string {
	return fmt.Sprintf(`
resource "materialize_connection_kafka" "test" {
	name = "%s"
	kafka_broker {
		broker = "redpanda:9092"
	}
}

resource "materialize_table" "test" {
	name = "%s"
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
	name = "%s"
	kafka_connection {
		name = materialize_connection_kafka.test.name
	}
	from {
		name = materialize_table.test.name
	}
	size  = "1"
	topic = "sink_topic"
	format {
		json = true
	}
	envelope {
		debezium = true
	}
}
`, connName, tableName, sinkName)
}

func testAccCheckSinkKafkaExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("sink kafka not found: %s", name)
		}
		_, err := materialize.ScanSink(db, r.Primary.ID)
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

		_, err := materialize.ScanSink(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("sink %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

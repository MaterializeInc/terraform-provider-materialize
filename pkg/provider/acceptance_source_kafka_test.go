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

func TestAccSourceKafka_basic(t *testing.T) {
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceKafkaResource(connName, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceKafkaExists("materialize_source_kafka.test"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, sourceName)),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "topic", "topic1"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "key_format.0.text", "true"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "value_format.0.text", "true"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "envelope.0.none", "true"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "envelope.0.debezium", "false"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "envelope.0.upsert", "false"),
				),
			},
		},
	})
}

func TestAccSourceKafka_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	connName := fmt.Sprintf("conn_%s", slug)

	sourceName := fmt.Sprintf("old_%s", slug)
	newSourceName := fmt.Sprintf("new_%s", slug)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceKafkaResource(connName, sourceName),
			},
			{
				Config: testAccSourceKafkaResource(connName, newSourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceKafkaExists("materialize_source_kafka.test"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "name", newSourceName),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newSourceName)),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "topic", "topic1"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "key_format.0.text", "true"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "value_format.0.text", "true"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "envelope.0.none", "true"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "envelope.0.debezium", "false"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "envelope.0.upsert", "false"),
				),
			},
		},
	})
}

func TestAccSourceKafka_disappears(t *testing.T) {
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceKafkaDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceKafkaResource(connName, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceKafkaExists("materialize_source_kafka.test"),
					testAccCheckSourceKafkaDisappears(sourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSourceKafkaResource(connName string, sourceName string) string {
	return fmt.Sprintf(`
resource "materialize_connection_kafka" "test" {
	name = "%s"
	kafka_broker {
		broker = "redpanda:9092"
	}
}

resource "materialize_source_kafka" "test" {
	name = "%s"
	kafka_connection {
		name = materialize_connection_kafka.test.name
	}

	size  = "1"
	topic = "topic1"
	key_format {
		text = true
	}
	value_format {
		text = true
	}
	envelope {
		none = true
	}
}
`, connName, sourceName)
}

func testAccCheckSourceKafkaExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("source kafka not found: %s", name)
		}
		_, err := materialize.ScanSource(db, r.Primary.ID)
		return err
	}
}

func testAccCheckSourceKafkaDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP SOURCE "%s";`, name))
		return err
	}
}

func testAccCheckAllSourceKafkaDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_source_kafka" {
			continue
		}

		_, err := materialize.ScanSource(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("source %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

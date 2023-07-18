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
				Config: testAccSourceKafkaResource(roleName, connName, sourceName, source2Name, roleName),
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
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "ownership_role", "mz_system"),
					testAccCheckSourceKafkaExists("materialize_source_kafka.test_role"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test_role", "name", source2Name),
					resource.TestCheckResourceAttr("materialize_source_kafka.test_role", "ownership_role", roleName),
				),
			},
		},
	})
}

func TestAccSourceKafka_update(t *testing.T) {
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
				Config: testAccSourceKafkaResource(roleName, connName, sourceName, source2Name, "mz_system"),
			},
			{
				Config: testAccSourceKafkaResource(roleName, connName, newSourceName, source2Name, roleName),
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
					testAccCheckSourceKafkaExists("materialize_source_kafka.test_role"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test_role", "ownership_role", roleName),
				),
			},
		},
	})
}

func TestAccSourceKafka_disappears(t *testing.T) {
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
				Config: testAccSourceKafkaResource(roleName, connName, sourceName, source2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceKafkaExists("materialize_source_kafka.test"),
					testAccCheckObjectDisappears(
						materialize.ObjectSchemaStruct{
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

func testAccSourceKafkaResource(roleName, connName, sourceName, source2Name, sourceOwner string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_connection_kafka" "test" {
	name = "%[2]s"
	kafka_broker {
		broker = "redpanda:9092"
	}
}

resource "materialize_source_kafka" "test" {
	name = "%[3]s"
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

resource "materialize_source_kafka" "test_role" {
	name = "%[4]s"
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
	ownership_role = "%[5]s"

	depends_on = [materialize_role.test]
}
`, roleName, connName, sourceName, source2Name, sourceOwner)
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

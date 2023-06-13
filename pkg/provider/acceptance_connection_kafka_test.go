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

func TestAccConnKafka_basic(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.test"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
				),
			},
		},
	})
}

func TestAccConnKafka_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	connectionName := fmt.Sprintf("old_%s", slug)
	newConnectionName := fmt.Sprintf("new_%s", slug)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaResource(connectionName),
			},
			{
				Config: testAccConnKafkaResource(newConnectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.test"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "name", newConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newConnectionName)),
				),
			},
		},
	})
}

func TestAccConnKafka_disappears(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllConnKafkasDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnKafkaResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnKafkaExists("materialize_connection_kafka.test"),
					testAccCheckConnKafkaDisappears(connectionName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConnKafkaResource(name string) string {
	return fmt.Sprintf(`
resource "materialize_connection_kafka" "test" {
	name = "%s"
	kafka_broker {
		broker = "redpanda:9092"
	}
}
`, name)
}

func testAccCheckConnKafkaExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("connection kafka not found: %s", name)
		}
		_, err := materialize.ScanConnection(db, r.Primary.ID)
		return err
	}
}

func testAccCheckConnKafkaDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP CONNECTION "%s";`, name))
		return err
	}
}

func testAccCheckAllConnKafkasDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_connection_kafka" {
			continue
		}

		_, err := materialize.ScanConnection(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("connection %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

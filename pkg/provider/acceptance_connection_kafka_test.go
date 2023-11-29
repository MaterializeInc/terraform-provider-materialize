package provider

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jmoiron/sqlx"
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
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.#", "1"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "kafka_broker.0.broker", "redpanda:9092"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_connection_kafka.test", "comment", "object comment"),
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
							ObjectType: "CONNECTION",
							Name:       connectionName,
						},
					),
				),
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

func testAccCheckAllConnKafkaDestroyed(s *terraform.State) error {
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

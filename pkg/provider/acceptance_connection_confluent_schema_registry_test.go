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

func TestAccConnConfluentSchemaRegistry_basic(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnConfluentSchemaRegistryResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnConfluentSchemaRegistryExists("materialize_connection_confluent_schema_registry.test"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "url", "http://redpanda:8081"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
				),
			},
		},
	})
}

func TestAccConnConfluentSchemaRegistry_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	connectionName := fmt.Sprintf("old_%s", slug)
	newConnectionName := fmt.Sprintf("new_%s", slug)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnConfluentSchemaRegistryResource(connectionName),
			},
			{
				Config: testAccConnConfluentSchemaRegistryResource(newConnectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnConfluentSchemaRegistryExists("materialize_connection_confluent_schema_registry.test"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "name", newConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newConnectionName)),
				),
			},
		},
	})
}

func TestAccConnConfluentSchemaRegistry_disappears(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllConnConfluentSchemaRegistryDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnConfluentSchemaRegistryResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnConfluentSchemaRegistryExists("materialize_connection_confluent_schema_registry.test"),
					testAccCheckConnConfluentSchemaRegistryDisappears(connectionName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConnConfluentSchemaRegistryResource(name string) string {
	return fmt.Sprintf(`
resource "materialize_connection_confluent_schema_registry" "test" {
	name = "%s"
	url  = "http://redpanda:8081"
}
`, name)
}

func testAccCheckConnConfluentSchemaRegistryExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("connection confluent schema registry not found: %s", name)
		}
		_, err := materialize.ScanConnection(db, r.Primary.ID)
		return err
	}
}

func testAccCheckConnConfluentSchemaRegistryDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP CONNECTION "%s";`, name))
		return err
	}
}

func testAccCheckAllConnConfluentSchemaRegistryDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_connection_confluent_schema_registry" {
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

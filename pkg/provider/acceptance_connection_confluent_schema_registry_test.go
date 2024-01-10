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

func TestAccConnConfluentSchemaRegistry_basic(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnConfluentSchemaRegistryResource(roleName, connectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnConfluentSchemaRegistryExists("materialize_connection_confluent_schema_registry.test"),
					resource.TestMatchResourceAttr("materialize_connection_confluent_schema_registry.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "url", "http://redpanda:8081"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "ownership_role", "mz_system"),
					testAccCheckConnConfluentSchemaRegistryExists("materialize_connection_confluent_schema_registry.test_role"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test_role", "name", connection2Name),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test_role", "ownership_role", roleName),
				),
			},
			{
				ResourceName:      "materialize_connection_confluent_schema_registry.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccConnConfluentSchemaRegistry_update(t *testing.T) {
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
				Config: testAccConnConfluentSchemaRegistryResource(roleName, connectionName, connection2Name, "mz_system"),
			},
			{
				Config: testAccConnConfluentSchemaRegistryResource(roleName, newConnectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnConfluentSchemaRegistryExists("materialize_connection_confluent_schema_registry.test"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "name", newConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newConnectionName)),
					testAccCheckConnConfluentSchemaRegistryExists("materialize_connection_confluent_schema_registry.test_role"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.test_role", "ownership_role", roleName),
				),
			},
		},
	})
}

func TestAccConnConfluentSchemaRegistry_disappears(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllConnConfluentSchemaRegistryDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnConfluentSchemaRegistryResource(roleName, connectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnConfluentSchemaRegistryExists("materialize_connection_confluent_schema_registry.test"),
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

func testAccConnConfluentSchemaRegistryResource(roleName, connectionName, connection2Name, connectionOwner string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_connection_confluent_schema_registry" "test" {
	name = "%[2]s"
	url  = "http://redpanda:8081"
}

resource "materialize_connection_confluent_schema_registry" "test_role" {
	name = "%[3]s"
	url  = "http://redpanda:8081"
	ownership_role = "%[4]s"

	depends_on = [materialize_role.test]
}
`, roleName, connectionName, connection2Name, connectionOwner)
}

func testAccCheckConnConfluentSchemaRegistryExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("connection confluent schema registry not found: %s", name)
		}
		_, err = materialize.ScanConnection(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllConnConfluentSchemaRegistryDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_connection_confluent_schema_registry" {
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

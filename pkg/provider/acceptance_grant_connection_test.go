package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jmoiron/sqlx"
)

func TestAccGrantConnection_basic(t *testing.T) {
	privilege := randomPrivilege("CONNECTION")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantConnectionResource(roleName, connectionName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantConnectionExists("materialize_grant_connection.connection_grant", roleName, connectionName, schemaName, databaseName, privilege),
					resource.TestCheckResourceAttr("materialize_grant_connection.connection_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_grant_connection.connection_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_grant_connection.connection_grant", "connection_name", connectionName),
					resource.TestCheckResourceAttr("materialize_grant_connection.connection_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_grant_connection.connection_grant", "database_name", databaseName),
				),
			},
		},
	})
}

func TestAccGrantConnection_disappears(t *testing.T) {
	privilege := randomPrivilege("CONNECTION")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantConnectionResource(roleName, connectionName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantConnectionExists("materialize_grant_connection.connection_grant", roleName, connectionName, schemaName, databaseName, privilege),
					testAccCheckGrantConnectionRevoked(roleName, connectionName, schemaName, databaseName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantConnectionResource(roleName, connectionName, schemaName, databaseName, privilege string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%s"
}

resource "materialize_database" "test" {
	name = "%s"
}

resource "materialize_schema" "test" {
	name = "%s"
	database_name = materialize_database.test.name
}

resource "materialize_connection_kafka" "test" {
	name = "%s"
	schema_name   = materialize_schema.test.name
	database_name = materialize_database.test.name

	kafka_broker {
	  broker = "redpanda:9092"
	}
}

resource "materialize_grant_connection" "connection_grant" {
	role_name       = materialize_role.test.name
	privilege       = "%s"
	database_name   = materialize_database.test.name
	schema_name     = materialize_schema.test.name
	connection_name = materialize_connection_kafka.test.name
}
`, roleName, databaseName, schemaName, connectionName, privilege)
}

func testAccCheckGrantConnectionExists(grantName, roleName, connectionName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("grant not found")
		}

		o := materialize.ObjectSchemaStruct{Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
		id, err := materialize.ConnectionId(db, o)
		if err != nil {
			return err
		}

		roleId, err := materialize.RoleId(db, roleName)
		if err != nil {
			return err
		}

		g, err := materialize.ScanPrivileges(db, "CONNECTION", id)
		if err != nil {
			return err
		}

		privilegeMap := materialize.ParsePrivileges(g)
		if !materialize.HasPrivilege(privilegeMap[roleId], privilege) {
			return fmt.Errorf("connection object %s does not include privilege %s", g, privilege)
		}
		return nil
	}
}

func testAccCheckGrantConnectionRevoked(roleName, connectionName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`REVOKE %s ON CONNECTION "%s"."%s"."%s" FROM "%s";`, privilege, databaseName, schemaName, connectionName, roleName))
		return err
	}
}

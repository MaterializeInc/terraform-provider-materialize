package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					testAccCheckGrantExists(
						materialize.MaterializeObject{
							ObjectType:   "CONNECTION",
							Name:         connectionName,
							SchemaName:   schemaName,
							DatabaseName: databaseName,
						},
						"materialize_connection_grant.connection_grant", roleName, privilege),
					resource.TestMatchResourceAttr("materialize_connection_grant.connection_grant", "id", terraformGrantIdRegex),
					resource.TestCheckResourceAttr("materialize_connection_grant.connection_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_connection_grant.connection_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_connection_grant.connection_grant", "connection_name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_grant.connection_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_connection_grant.connection_grant", "database_name", databaseName),
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

	o := materialize.MaterializeObject{
		ObjectType:   "CONNECTION",
		Name:         connectionName,
		SchemaName:   schemaName,
		DatabaseName: databaseName,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantConnectionResource(roleName, connectionName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(o, "materialize_connection_grant.connection_grant", roleName, privilege),
					testAccCheckGrantRevoked(o, roleName, privilege),
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
	security_protocol = "PLAINTEXT"
}

resource "materialize_connection_grant" "connection_grant" {
	role_name       = materialize_role.test.name
	privilege       = "%s"
	database_name   = materialize_database.test.name
	schema_name     = materialize_schema.test.name
	connection_name = materialize_connection_kafka.test.name
}
`, roleName, databaseName, schemaName, connectionName, privilege)
}

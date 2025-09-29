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

func TestAccConnectionSQLServer_basic(t *testing.T) {
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionSQLServerResource(roleName, secretName, connectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionSQLServerExists("materialize_connection_sqlserver.test"),
					resource.TestMatchResourceAttr("materialize_connection_sqlserver.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "user.#", "1"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "user.0.text", "sa"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "password.#", "1"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "password.0.name", secretName),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "password.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "password.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "database", "testdb"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "comment", "object comment"),
					testAccCheckConnectionSQLServerExists("materialize_connection_sqlserver.test_role"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test_role", "name", connection2Name),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test_role", "ownership_role", roleName),
				),
			},
			{
				ResourceName:      "materialize_connection_sqlserver.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccConnectionSQLServer_update(t *testing.T) {
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
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
				Config: testAccConnectionSQLServerResource(roleName, secretName, connectionName, connection2Name, "mz_system"),
			},
			{
				Config: testAccConnectionSQLServerResource(roleName, secretName, newConnectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionSQLServerExists("materialize_connection_sqlserver.test"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "name", newConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newConnectionName)),
					testAccCheckConnectionSQLServerExists("materialize_connection_sqlserver.test_role"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test_role", "ownership_role", roleName),
				),
			},
		},
	})
}

func TestAccConnectionSQLServer_disappears(t *testing.T) {
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllConnectionSQLServerDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionSQLServerResource(roleName, secretName, connectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionSQLServerExists("materialize_connection_sqlserver.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "CONNECTION",
							Name:       connectionName,
						},
					),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccConnectionSQLServer_updateConnectionAttributes(t *testing.T) {
	initialSecretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updatedSecretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	initialConnectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updatedConnectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	initialHost := "initial_host"
	updatedHost := "updated_host"
	initialPort := "1433"
	updatedPort := "1434"
	initialDatabase := "initial_database"
	updatedDatabase := "updated_database"
	sshTunnelName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	sshTunnel2Name := sshTunnelName + "_2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionSQLServerResourceUpdates(roleName, initialSecretName, initialConnectionName, initialHost, initialPort, sshTunnelName, initialDatabase),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionSQLServerExists("materialize_connection_sqlserver.test"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "name", initialConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "host", initialHost),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "port", initialPort),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "database", initialDatabase),
				),
			},
			{
				Config: testAccConnectionSQLServerResourceUpdates(roleName, updatedSecretName, updatedConnectionName, updatedHost, updatedPort, sshTunnel2Name, updatedDatabase),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionSQLServerExists("materialize_connection_sqlserver.test"),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "name", updatedConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "host", updatedHost),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "port", updatedPort),
					resource.TestCheckResourceAttr("materialize_connection_sqlserver.test", "database", updatedDatabase),
				),
			},
		},
	})
}

func testAccConnectionSQLServerResource(roleName, secretName, connectionName, connection2Name, connectionOwner string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_secret" "sqlserver_password" {
	name          = "%[2]s"
	value         = "Password123!"
}

resource "materialize_connection_sqlserver" "test" {
	name = "%[3]s"
	host = "sqlserver"
	port = 1433
	user {
		text = "sa"
	}
	password {
		name          = materialize_secret.sqlserver_password.name
		schema_name   = materialize_secret.sqlserver_password.schema_name
		database_name = materialize_secret.sqlserver_password.database_name
	}
	database = "testdb"
	comment  = "object comment"
}

resource "materialize_connection_sqlserver" "test_role" {
	name = "%[4]s"
	host = "sqlserver"
	port = 1433
	user {
		text = "sa"
	}
	password {
		name          = materialize_secret.sqlserver_password.name
		schema_name   = materialize_secret.sqlserver_password.schema_name
		database_name = materialize_secret.sqlserver_password.database_name
	}
	database = "testdb"
	ownership_role = "%[5]s"

	depends_on = [materialize_role.test]
}
`, roleName, secretName, connectionName, connection2Name, connectionOwner)
}

func testAccConnectionSQLServerResourceUpdates(roleName, secretName, connectionName, host, port, sshTunnelName, database string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "test" {
		name = "%[1]s"
	}

	resource "materialize_secret" "sqlserver_password" {
		name  = "%[2]s"
		value = "Password123!"
	}

	resource "materialize_connection_ssh_tunnel" "ssh_connection" {
		name        = "%[6]s"
		schema_name = "public"
		comment     = "connection ssh tunnel comment"

		host = "ssh_host"
		user = "ssh_user"
		port = 22

		validate = false
	}

	resource "materialize_connection_sqlserver" "test" {
		name           = "%[3]s"
		host           = "%[4]s"
		port           = %[5]s
		user {
			text = "sa"
		}
		password {
			name          = materialize_secret.sqlserver_password.name
			schema_name   = materialize_secret.sqlserver_password.schema_name
			database_name = materialize_secret.sqlserver_password.database_name
		}
		ssh_tunnel {
			name = materialize_connection_ssh_tunnel.ssh_connection.name
		}
		database       = "%[7]s"

		comment        = "Test connection"
		ownership_role = materialize_role.test.name
		validate	   = false
	}
	`, roleName, secretName, connectionName, host, port, sshTunnelName, database)
}

func testAccCheckConnectionSQLServerExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("connection sqlserver not found: %s", name)
		}
		_, err = materialize.ScanConnection(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllConnectionSQLServerDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_connection_sqlserver" {
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

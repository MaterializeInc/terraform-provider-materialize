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

func TestAccConnectionMySQL_basic(t *testing.T) {
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
				Config: testAccConnectionMySQLResource(roleName, secretName, connectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionMySQLExists("materialize_connection_mysql.test"),
					resource.TestMatchResourceAttr("materialize_connection_mysql.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_connection_mysql.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_mysql.test", "user.#", "1"),
					resource.TestCheckResourceAttr("materialize_connection_mysql.test", "password.#", "1"),
					resource.TestCheckResourceAttr("materialize_connection_mysql.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_mysql.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_mysql.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
					resource.TestCheckResourceAttr("materialize_connection_mysql.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_connection_mysql.test", "comment", "object comment"),
					testAccCheckConnectionMySQLExists("materialize_connection_mysql.test_role"),
					resource.TestCheckResourceAttr("materialize_connection_mysql.test_role", "name", connection2Name),
					resource.TestCheckResourceAttr("materialize_connection_mysql.test_role", "ownership_role", roleName),
				),
			},
			{
				ResourceName:      "materialize_connection_mysql.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccConnectionMySQL_update(t *testing.T) {
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
				Config: testAccConnectionMySQLResource(roleName, secretName, connectionName, connection2Name, "mz_system"),
			},
			{
				Config: testAccConnectionMySQLResource(roleName, secretName, newConnectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionMySQLExists("materialize_connection_mysql.test"),
					resource.TestCheckResourceAttr("materialize_connection_mysql.test", "name", newConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_mysql.test_role", "ownership_role", roleName),
				),
			},
		},
	})
}

func TestAccConnectionMySQL_disappears(t *testing.T) {
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllConnectionMySQLDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionMySQLResource(roleName, secretName, connectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionMySQLExists("materialize_connection_mysql.test"),
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

func TestAccConnectionMySQL_updateAttributes(t *testing.T) {
	initialSecretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updatedSecretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	initialConnectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updatedConnectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	initialHost := "initial_mysql_host"
	updatedHost := "updated_mysql_host"
	initialPort := 3306
	updatedPort := 3307
	initialSslMode := "verify-ca"
	updatedSslMode := "required"
	initialSslCa := "initial_ssl_ca"
	updatedSslCa := "updated_ssl_ca"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllConnKafkaDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionMySQLResourceUpdates(roleName, initialSecretName, initialConnectionName, initialHost, initialPort, initialSslMode, initialSslCa),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionMySQLExists("materialize_connection_mysql.mysql_update"),
					resource.TestCheckResourceAttr("materialize_connection_mysql.mysql_update", "name", initialConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_mysql.mysql_update", "host", initialHost),
					resource.TestCheckResourceAttr("materialize_connection_mysql.mysql_update", "port", fmt.Sprintf("%d", initialPort)),
					resource.TestCheckResourceAttr("materialize_connection_mysql.mysql_update", "ssl_mode", initialSslMode),
					resource.TestCheckResourceAttr("materialize_connection_mysql.mysql_update", "ssl_certificate_authority.0.text", initialSslCa),
				),
			},
			{
				Config: testAccConnectionMySQLResourceUpdates(roleName, updatedSecretName, updatedConnectionName, updatedHost, updatedPort, updatedSslMode, updatedSslCa),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionMySQLExists("materialize_connection_mysql.mysql_update"),
					resource.TestCheckResourceAttr("materialize_connection_mysql.mysql_update", "name", updatedConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_mysql.mysql_update", "host", updatedHost),
					resource.TestCheckResourceAttr("materialize_connection_mysql.mysql_update", "port", fmt.Sprintf("%d", updatedPort)),
					resource.TestCheckResourceAttr("materialize_connection_mysql.mysql_update", "ssl_mode", updatedSslMode),
					resource.TestCheckResourceAttr("materialize_connection_mysql.mysql_update", "ssl_certificate_authority.0.text", updatedSslCa),
				),
			},
		},
	})
}

func testAccCheckAllConnectionMySQLDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_connection_mysql" {
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

func testAccConnectionMySQLResource(roleName, secretName, connectionName, connection2Name, connectionOwner string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_secret" "mysql_password" {
	name          = "%[2]s"
	value         = "c2VjcmV0Cg=="
}

resource "materialize_connection_mysql" "test" {
	name = "%[3]s"
	host = "mysql"
	port = 3306
	user {
		text = "mysqluser"
	}
	password {
		name          = materialize_secret.mysql_password.name
		schema_name   = materialize_secret.mysql_password.schema_name
		database_name = materialize_secret.mysql_password.database_name
	}
	comment  = "object comment"
}

resource "materialize_connection_mysql" "test_role" {
	name = "%[4]s"
	host = "mysql"
	port = 3306
	user {
		text = "mysqluser"
	}
	password {
		name          = materialize_secret.mysql_password.name
		schema_name   = materialize_secret.mysql_password.schema_name
		database_name = materialize_secret.mysql_password.database_name
	}
	ownership_role = "%[5]s"

	depends_on = [materialize_role.test]
}
`, roleName, secretName, connectionName, connection2Name, connectionOwner)
}

func testAccConnectionMySQLResourceUpdates(roleName, secretName, connectionName, host string, port int, sslMode, ca string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_secret" "mysql_password" {
	name  = "%[2]s"
	value = "c2VjcmV0Cg=="
}

resource "materialize_connection_mysql" "mysql_update" {
	name           = "%[3]s"
	host           = "%[4]s"
	port           = %[5]d
	user {
		text = "mysqluser"
	}
	password {
		name          = materialize_secret.mysql_password.name
		schema_name   = materialize_secret.mysql_password.schema_name
		database_name = materialize_secret.mysql_password.database_name
	}
	ssl_mode       = "%[6]s"
	ssl_certificate_authority {
		text = "%[7]s"
	}

	comment        = "Test connection"
	ownership_role = materialize_role.test.name
	validate       = false
}
`, roleName, secretName, connectionName, host, port, sslMode, ca)
}

func testAccCheckConnectionMySQLExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("connection MySQL not found: %s", name)
		}
		_, err = materialize.ScanConnection(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

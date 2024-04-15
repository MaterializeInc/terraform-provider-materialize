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
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccConnConfluentSchemaRegistry_updateExtended(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	// Initial settings
	initialURL := "http://initial-redpanda:8081"
	initialUsername := "initial_user"
	initialPasswordName := "initial_password"
	initialSSLKey := "initial_ssl_key"
	initialSSLCert := "initial_ssl_cert"
	initialSSLCA := "initial_ca_cert"

	// Updated settings
	updatedURL := "http://updated-redpanda:8081"
	updatedUsername := "updated_user"
	updatedPasswordName := "updated_password"
	updatedSSLKey := "updated_ssl_key"
	updatedSSLCert := "updated_ssl_cert"
	updatedSSLCA := "updated_ca_cert"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllConnConfluentSchemaRegistryDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnConfluentSchemaRegistryResourceUpdates(connectionName, initialURL, initialUsername, initialPasswordName, initialSSLCert, initialSSLKey, initialSSLCA),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnConfluentSchemaRegistryExists("materialize_connection_confluent_schema_registry.csr_updates"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.csr_updates", "url", initialURL),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.csr_updates", "username.0.text", initialUsername),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.csr_updates", "password.0.name", initialPasswordName),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.csr_updates", "ssl_certificate.0.text", initialSSLCert),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.csr_updates", "ssl_key.0.name", initialSSLKey),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.csr_updates", "ssl_certificate_authority.0.text", initialSSLCA),
				),
			},
			{
				Config: testAccConnConfluentSchemaRegistryResourceUpdates(connectionName, updatedURL, updatedUsername, updatedPasswordName, updatedSSLCert, updatedSSLKey, updatedSSLCA),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnConfluentSchemaRegistryExists("materialize_connection_confluent_schema_registry.csr_updates"),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.csr_updates", "url", updatedURL),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.csr_updates", "username.0.text", updatedUsername),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.csr_updates", "password.0.name", updatedPasswordName),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.csr_updates", "ssl_certificate.0.text", updatedSSLCert),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.csr_updates", "ssl_key.0.name", updatedSSLKey),
					resource.TestCheckResourceAttr("materialize_connection_confluent_schema_registry.csr_updates", "ssl_certificate_authority.0.text", updatedSSLCA),
				),
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

func testAccConnConfluentSchemaRegistryResourceUpdates(connectionName, url, usernameText, passwordName, sslCertText, sslKeyName, sslCAText string) string {
	return fmt.Sprintf(`
resource "materialize_secret" "csr_password2" {
	name  = "%[4]s"
	value = "csr_password"
}

resource "materialize_secret" "ssl_key2" {
    name  = "%[6]s"
	value = "ssl_key"
}

resource "materialize_connection_confluent_schema_registry" "csr_updates" {
    name                      = "%[1]s"
    url                       = "%[2]s"

    username {
        text = "%[3]s"
    }

    password {
        name          = materialize_secret.csr_password2.name
        database_name = materialize_secret.csr_password2.database_name
        schema_name   = materialize_secret.csr_password2.schema_name
    }

    ssl_certificate {
        text = "%[5]s"
    }

    ssl_key {
        name          = materialize_secret.ssl_key2.name
        database_name = materialize_secret.ssl_key2.database_name
        schema_name   = materialize_secret.ssl_key2.schema_name
    }

    ssl_certificate_authority {
        text = "%[7]s"
    }

    validate = false
}
`, connectionName, url, usernameText, passwordName, sslCertText, sslKeyName, sslCAText)
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

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
	"github.com/jmoiron/sqlx"
)

func TestAccConnPostgres_basic(t *testing.T) {
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
				Config: testAccConnPostgresResource(roleName, secretName, connectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnPostgresExists("materialize_connection_postgres.test"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "user.#", "1"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "user.0.text", "postgres"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "password.#", "1"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "password.0.name", secretName),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "password.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "password.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "database", "postgres"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "comment", "object comment"),
					testAccCheckConnPostgresExists("materialize_connection_postgres.test_role"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test_role", "name", connection2Name),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test_role", "ownership_role", roleName),
				),
			},
			{
				ResourceName:      "materialize_connection_postgres.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccConnPostgres_update(t *testing.T) {
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
				Config: testAccConnPostgresResource(roleName, secretName, connectionName, connection2Name, "mz_system"),
			},
			{
				Config: testAccConnPostgresResource(roleName, secretName, newConnectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnPostgresExists("materialize_connection_postgres.test"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "name", newConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newConnectionName)),
					testAccCheckConnPostgresExists("materialize_connection_postgres.test_role"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test_role", "ownership_role", roleName),
				),
			},
		},
	})
}

func TestAccConnPostgres_disappears(t *testing.T) {
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllConnPostgresDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnPostgresResource(roleName, secretName, connectionName, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnPostgresExists("materialize_connection_postgres.test"),
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

func testAccConnPostgresResource(roleName, secretName, connectionName, connection2Name, connectionOwner string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_secret" "postgres_password" {
	name          = "%[2]s"
	value         = "c2VjcmV0Cg=="
}

resource "materialize_connection_postgres" "test" {
	name = "%[3]s"
	host = "postgres"
	port = 5432
	user {
		text = "postgres"
	}
	password {
		name          = materialize_secret.postgres_password.name
		schema_name   = materialize_secret.postgres_password.schema_name
		database_name = materialize_secret.postgres_password.database_name
	}
	database = "postgres"
	comment  = "object comment"
}

resource "materialize_connection_postgres" "test_role" {
	name = "%[4]s"
	host = "postgres"
	port = 5432
	user {
		text = "postgres"
	}
	password {
		name          = materialize_secret.postgres_password.name
		schema_name   = materialize_secret.postgres_password.schema_name
		database_name = materialize_secret.postgres_password.database_name
	}
	database = "postgres"
	ownership_role = "%[5]s"

	depends_on = [materialize_role.test]
}
`, roleName, secretName, connectionName, connection2Name, connectionOwner)
}

func testAccCheckConnPostgresExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("connection postgres not found: %s", name)
		}
		_, err := materialize.ScanConnection(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllConnPostgresDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_connection_postgres" {
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

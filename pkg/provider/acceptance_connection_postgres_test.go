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

func TestAccConnPostgres_basic(t *testing.T) {
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnPostgresResource(secretName, connectionName),
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
				),
			},
		},
	})
}

func TestAccConnPostgres_update(t *testing.T) {
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	connectionName := fmt.Sprintf("old_%s", slug)
	newConnectionName := fmt.Sprintf("new_%s", slug)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnPostgresResource(secretName, connectionName),
			},
			{
				Config: testAccConnPostgresResource(secretName, newConnectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnPostgresExists("materialize_connection_postgres.test"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "name", newConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_postgres.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newConnectionName)),
				),
			},
		},
	})
}

func TestAccConnPostgres_disappears(t *testing.T) {
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllConnPostgresDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnPostgresResource(secretName, connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnPostgresExists("materialize_connection_postgres.test"),
					testAccCheckConnPostgresDisappears(connectionName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConnPostgresResource(secret, name string) string {
	return fmt.Sprintf(`
resource "materialize_secret" "postgres_password" {
	name          = "%s"
	value         = "c2VjcmV0Cg=="
}

resource "materialize_connection_postgres" "test" {
	name = "%s"
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
}
`, secret, name)
}

func testAccCheckConnPostgresExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("connection postgres not found: %s", name)
		}
		_, err := materialize.ScanConnection(db, r.Primary.ID)
		return err
	}
}

func testAccCheckConnPostgresDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP CONNECTION "%s";`, name))
		return err
	}
}

func testAccCheckAllConnPostgresDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_connection_postgres" {
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

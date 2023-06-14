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

func TestAccSourcePostgres_basic(t *testing.T) {
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourcePostgresResource(secretName, connName, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourcePostgresExists("materialize_source_postgres.test"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, sourceName)),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "size", "1"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.#", "2"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.0.name", "table1"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.0.alias", fmt.Sprintf(`%s_table1`, connName)),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.1.name", "table2"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.1.alias", fmt.Sprintf(`%s_table2`, connName)),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "publication", "mz_source"),
				),
			},
		},
	})
}

func TestAccSourcePostgres_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	secretName := fmt.Sprintf("secret_%s", slug)
	connName := fmt.Sprintf("conn_%s", slug)

	sourceName := fmt.Sprintf("old_%s", slug)
	newSourceName := fmt.Sprintf("new_%s", slug)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourcePostgresResource(secretName, connName, sourceName),
			},
			{
				Config: testAccSourcePostgresResource(secretName, connName, newSourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourcePostgresExists("materialize_source_postgres.test"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "name", newSourceName),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newSourceName)),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "size", "1"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.#", "2"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.0.name", "table1"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.0.alias", fmt.Sprintf(`%s_table1`, connName)),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.1.name", "table2"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.1.alias", fmt.Sprintf(`%s_table2`, connName)),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "publication", "mz_source"),
				),
			},
		},
	})
}

func TestAccSourcePostgres_disappears(t *testing.T) {
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourcePostgresDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourcePostgresResource(secretName, connName, sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourcePostgresExists("materialize_source_postgres.test"),
					testAccCheckSourcePostgresDisappears(sourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSourcePostgresResource(secretName string, connName string, sourceName string) string {
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

resource "materialize_source_postgres" "test" {
	name = "%s"
	postgres_connection {
		name = materialize_connection_postgres.test.name
	}

	size  = "1"
	publication = "mz_source"
	table {
	  name  = "table1"
	  alias = "%[2]s_table1"
	}
	table {
	  name  = "table2"
	  alias = "%[2]s_table2"
	}
}
`, secretName, connName, sourceName)
}

func testAccCheckSourcePostgresExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("source postgres not found: %s", name)
		}
		_, err := materialize.ScanSource(db, r.Primary.ID)
		return err
	}
}

func testAccCheckSourcePostgresDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP SOURCE "%s";`, name))
		return err
	}
}

func testAccCheckAllSourcePostgresDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_source_postgres" {
			continue
		}

		_, err := materialize.ScanSource(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("source %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

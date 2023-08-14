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
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	source2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourcePostgresResource(roleName, secretName, connName, sourceName, source2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourcePostgresExists("materialize_source_postgres.test"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, sourceName)),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "size", "1"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "text_columns.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.#", "2"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.0.name", "table1"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.0.alias", fmt.Sprintf(`%s_table1`, connName)),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.1.name", "table2"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.1.alias", fmt.Sprintf(`%s_table2`, connName)),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "publication", "mz_source"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "ownership_role", "mz_system"),
					testAccCheckSourcePostgresExists("materialize_source_postgres.test_role"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test_role", "name", source2Name),
					resource.TestCheckResourceAttr("materialize_source_postgres.test_role", "ownership_role", roleName),
				),
			},
			{
				ResourceName:      "materialize_source_postgres.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSourcePostgres_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	sourceName := fmt.Sprintf("old_%s", slug)
	newSourceName := fmt.Sprintf("new_%s", slug)
	source2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourcePostgresResource(roleName, secretName, connName, sourceName, source2Name, "mz_system"),
			},
			{
				Config: testAccSourcePostgresResourceUpdate(roleName, secretName, connName, newSourceName, source2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourcePostgresExists("materialize_source_postgres.test"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "name", newSourceName),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newSourceName)),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "size", "1"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "text_columns.#", "2"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.#", "2"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.0.name", "table1"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.0.alias", fmt.Sprintf(`%s_table1`, connName)),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.1.name", "table3"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.1.alias", fmt.Sprintf(`%s_table3`, connName)),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "publication", "mz_source"),
					testAccCheckSourcePostgresExists("materialize_source_postgres.test_role"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test_role", "ownership_role", roleName),
				),
			},
			{
				Config: testAccSourcePostgresResource(roleName, secretName, connName, newSourceName, source2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourcePostgresExists("materialize_source_postgres.test"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.#", "2"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.1.name", "table2"),
					resource.TestCheckResourceAttr("materialize_source_postgres.test", "table.1.alias", fmt.Sprintf(`%s_table2`, connName)),
				),
			},
		},
	})
}

func TestAccSourcePostgres_disappears(t *testing.T) {
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	source2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourcePostgresDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourcePostgresResource(roleName, secretName, connName, sourceName, source2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourcePostgresExists("materialize_source_postgres.test"),
					testAccCheckObjectDisappears(
						materialize.ObjectSchemaStruct{
							ObjectType: "SOURCE",
							Name:       sourceName,
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSourcePostgresResource(roleName, secretName, connName, sourceName, source2Name, sourceOwner string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_secret" "postgres_password" {
	name  = "%[2]s"
	value = "c2VjcmV0Cg=="
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
}

resource "materialize_source_postgres" "test" {
	name = "%[4]s"
	postgres_connection {
		name = materialize_connection_postgres.test.name
	}

	size  = "1"
	publication = "mz_source"
	table {
		name  = "table1"
		alias = "%[3]s_table1"
	}
	table {
		name  = "table2"
		alias = "%[3]s_table2"
	}
	text_columns = ["table1.id"]
}

resource "materialize_source_postgres" "test_role" {
	name = "%[5]s"
	postgres_connection {
		name = materialize_connection_postgres.test.name
	}

	size  = "1"
	publication = "mz_source"
	table {
		name  = "table1"
		alias = "%[3]s_table_role_1"
	}
	table {
		name  = "table2"
		alias = "%[3]s_table_role_2"
	}
	ownership_role = "%[6]s"

	depends_on = [materialize_role.test]
}
`, roleName, secretName, connName, sourceName, source2Name, sourceOwner)
}

func testAccSourcePostgresResourceUpdate(roleName, secretName, connName, sourceName, source2Name, sourceOwner string) string {
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
}

resource "materialize_source_postgres" "test" {
	name = "%[4]s"
	postgres_connection {
		name = materialize_connection_postgres.test.name
	}

	size  = "1"
	publication = "mz_source"
	table {
		name  = "table1"
		alias = "%[3]s_table1"
	}
	table {
		name  = "table3"
		alias = "%[3]s_table3"
	}
	text_columns = ["table1.id", "table3.id"]
}

resource "materialize_source_postgres" "test_role" {
	name = "%[5]s"
	postgres_connection {
		name = materialize_connection_postgres.test.name
	}

	size  = "1"
	publication = "mz_source"
	table {
		name  = "table1"
		alias = "%[3]s_table_role_1"
	}
	table {
		name  = "table2"
		alias = "%[3]s_table_role_2"
	}
	ownership_role = "%[6]s"

	depends_on = [materialize_role.test]
}
`, roleName, secretName, connName, sourceName, source2Name, sourceOwner)
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

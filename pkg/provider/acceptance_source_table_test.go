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

func TestAccSourceTablePostgres_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTablePostgresBasicResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table.test_postgres"),
					resource.TestMatchResourceAttr("materialize_source_table.test_postgres", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_table.test_postgres", "name", nameSpace+"_table_postgres"),
					resource.TestCheckResourceAttr("materialize_source_table.test_postgres", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_table.test_postgres", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table.test_postgres", "text_columns.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table.test_postgres", "text_columns.0", "updated_at"),
					resource.TestCheckResourceAttr("materialize_source_table.test_postgres", "upstream_name", "table2"),
					resource.TestCheckResourceAttr("materialize_source_table.test_postgres", "upstream_schema_name", "public"),
				),
			},
		},
	})
}

func TestAccSourceTableMySQL_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableMySQLBasicResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table.test_mysql"),
					resource.TestMatchResourceAttr("materialize_source_table.test_mysql", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_table.test_mysql", "name", nameSpace+"_table_mysql"),
					resource.TestCheckResourceAttr("materialize_source_table.test_mysql", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_table.test_mysql", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table.test_mysql", "upstream_name", "mysql_table1"),
					resource.TestCheckResourceAttr("materialize_source_table.test_mysql", "upstream_schema_name", "shop"),
				),
			},
		},
	})
}

func TestAccSourceTableLoadGen_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableLoadGenBasicResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table.test_loadgen"),
					resource.TestMatchResourceAttr("materialize_source_table.test_loadgen", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_table.test_loadgen", "name", nameSpace+"_table_loadgen"),
					resource.TestCheckResourceAttr("materialize_source_table.test_loadgen", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_table.test_loadgen", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table.test_loadgen", "upstream_name", "bids"),
				),
			},
		},
	})
}

func TestAccSourceTable_update(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableResource(nameSpace, "table2", "mz_system", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table.test"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "name", nameSpace+"_table"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "upstream_name", "table2"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "text_columns.#", "2"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "comment", ""),
				),
			},
			{
				Config: testAccSourceTableResource(nameSpace, "table3", nameSpace+"_role", "Updated comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table.test"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "name", nameSpace+"_table"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "upstream_name", "table3"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "text_columns.#", "2"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "ownership_role", nameSpace+"_role"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "comment", "Updated comment"),
				),
			},
		},
	})
}

func TestAccSourceTable_disappears(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceTableDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableResource(nameSpace, "table2", "mz_system", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "TABLE",
							Name:       nameSpace + "_table",
						},
					),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSourceTablePostgresBasicResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "postgres_password" {
		name  = "%[1]s_secret_postgres"
		value = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_postgres" "postgres_connection" {
		name    = "%[1]s_connection_postgres"
		// TODO: Change with container name once new image is available
		host    = "localhost"
		port    = 5432
		user {
			text = "postgres"
		}
		password {
			name = materialize_secret.postgres_password.name
		}
		database = "postgres"
	}

	resource "materialize_source_postgres" "test_source_postgres" {
		name         = "%[1]s_source_postgres"
		cluster_name = "quickstart"

		postgres_connection {
			name = materialize_connection_postgres.postgres_connection.name
		}
		publication = "mz_source"
		table {
			upstream_name  = "table2"
			upstream_schema_name = "public"
		}
	}

	resource "materialize_source_table" "test_postgres" {
		name           = "%[1]s_table_postgres"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name = materialize_source_postgres.test_source_postgres.name
		}

		upstream_name         = "table2"
		upstream_schema_name  = "public"

		text_columns = [
			"updated_at"
		]
	}
	`, nameSpace)
}

func testAccSourceTableMySQLBasicResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "mysql_password" {
		name  = "%[1]s_secret_mysql"
		value = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_mysql" "mysql_connection" {
		name    = "%[1]s_connection_mysql"
		// TODO: Change with container name once new image is available
		host    = "localhost"
		port    = 3306
		user {
			text = "repluser"
		}
		password {
			name = materialize_secret.mysql_password.name
		}
	}

	resource "materialize_source_mysql" "test_source_mysql" {
		name         = "%[1]s_source_mysql"
		cluster_name = "quickstart"

		mysql_connection {
			name = materialize_connection_mysql.mysql_connection.name
		}
		
		table {
			upstream_name        = "mysql_table1"
			upstream_schema_name = "shop"
			name                 = "mysql_table1_local"
		}
	}

	resource "materialize_source_table" "test_mysql" {
		name           = "%[1]s_table_mysql"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name = materialize_source_mysql.test_source_mysql.name
		}

		upstream_name         = "mysql_table1"
		upstream_schema_name  = "shop"
	}
	`, nameSpace)
}

func testAccSourceTableLoadGenBasicResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_source_load_generator" "test_loadgen" {
		name                = "%[1]s_loadgen"
		load_generator_type = "AUCTION"

		auction_options {
			tick_interval = "500ms"
		}
	}

	resource "materialize_source_table" "test_loadgen" {
		name           = "%[1]s_table_loadgen"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name = materialize_source_load_generator.test_loadgen.name
		}

		upstream_name = "bids"
	}
	`, nameSpace)
}

func testAccSourceTableBasicResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "postgres_password" {
		name  = "%[1]s_secret"
		value = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_postgres" "postgres_connection" {
		name    = "%[1]s_connection"
		// TODO: Change with container name once new image is available
		host    = "localhost"
		port    = 5432
		user {
			text = "postgres"
		}
		password {
			name          = materialize_secret.postgres_password.name
			database_name = materialize_secret.postgres_password.database_name
			schema_name   = materialize_secret.postgres_password.schema_name
		}
		database = "postgres"
	}

	resource "materialize_source_postgres" "test_source" {
		name         = "%[1]s_source"
		cluster_name = "quickstart"

		postgres_connection {
			name          = materialize_connection_postgres.postgres_connection.name
			schema_name   = materialize_connection_postgres.postgres_connection.schema_name
			database_name = materialize_connection_postgres.postgres_connection.database_name
		}
		publication = "mz_source"
		table {
			upstream_name  = "table2"
			upstream_schema_name = "public"
		}
	}

	resource "materialize_source_table" "test" {
		name           = "%[1]s_table"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name          = materialize_source_postgres.test_source.name
			schema_name   = "public"
			database_name = "materialize"
		}

		upstream_name         = "table2"
		upstream_schema_name  = "public"

		text_columns = [
			"updated_at"
		]
	}
	`, nameSpace)
}

func testAccSourceTableResource(nameSpace, upstreamName, ownershipRole, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "postgres_password" {
		name  = "%[1]s_secret"
		value = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_postgres" "postgres_connection" {
		name    = "%[1]s_connection"
		// TODO: Change with container name once new image is available
		host    = "localhost"
		port    = 5432
		user {
			text = "postgres"
		}
		password {
			name          = materialize_secret.postgres_password.name
			database_name = materialize_secret.postgres_password.database_name
			schema_name   = materialize_secret.postgres_password.schema_name
		}
		database = "postgres"
	}

	resource "materialize_source_postgres" "test_source" {
		name         = "%[1]s_source"
		cluster_name = "quickstart"

		postgres_connection {
			name          = materialize_connection_postgres.postgres_connection.name
			schema_name   = materialize_connection_postgres.postgres_connection.schema_name
			database_name = materialize_connection_postgres.postgres_connection.database_name
		}
		publication = "mz_source"
		table {
			upstream_name  = "%[2]s"
			upstream_schema_name = "public"
		}
	}

	resource "materialize_role" "test_role" {
		name = "%[1]s_role"
	}

	resource "materialize_source_table" "test" {
		name           = "%[1]s_table"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name          = materialize_source_postgres.test_source.name
			schema_name   = "public"
			database_name = "materialize"
		}

		upstream_name         = "%[2]s"
		upstream_schema_name  = "public"

		text_columns = [
			"updated_at",
			"id"
		]

		ownership_role = "%[3]s"
		comment        = "%[4]s"

		depends_on = [materialize_role.test_role]
	}
	`, nameSpace, upstreamName, ownershipRole, comment)
}

func testAccCheckSourceTableExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("source table not found: %s", name)
		}
		_, err = materialize.ScanSourceTable(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllSourceTableDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_source_table" {
			continue
		}

		_, err := materialize.ScanSourceTable(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("source table %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

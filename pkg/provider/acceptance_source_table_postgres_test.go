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
					testAccCheckSourceTablePostgresExists("materialize_source_table_postgres.test_postgres"),
					resource.TestMatchResourceAttr("materialize_source_table_postgres.test_postgres", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test_postgres", "name", nameSpace+"_table_postgres"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test_postgres", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test_postgres", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test_postgres", "text_columns.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test_postgres", "text_columns.0", "updated_at"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test_postgres", "upstream_name", "table2"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test_postgres", "upstream_schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test_postgres", "source.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test_postgres", "source.0.name", nameSpace+"_source_postgres"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test_postgres", "source.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test_postgres", "source.0.database_name", "materialize"),
				),
			},
			{
				ResourceName:      "materialize_source_table_postgres.test_postgres",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSourceTablePostgres_update(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTablePostgresResource(nameSpace, "table2", "mz_system", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTablePostgresExists("materialize_source_table_postgres.test"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "name", nameSpace+"_table"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "upstream_name", "table2"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "text_columns.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "comment", ""),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "source.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "source.0.name", nameSpace+"_source"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "source.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "source.0.database_name", "materialize"),
				),
			},
			{
				Config: testAccSourceTablePostgresResource(nameSpace, "table3", nameSpace+"_role", "Updated comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTablePostgresExists("materialize_source_table_postgres.test"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "name", nameSpace+"_table"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "upstream_name", "table3"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "text_columns.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "ownership_role", nameSpace+"_role"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "comment", "Updated comment"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "source.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "source.0.name", nameSpace+"_source"),
					resource.TestCheckResourceAttr("materialize_source_table_postgres.test", "source.0.schema_name", "public"),
				),
			},
		},
	})
}

func TestAccSourceTablePostgres_disappears(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceTablePostgresDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTablePostgresResource(nameSpace, "table2", "mz_system", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTablePostgresExists("materialize_source_table_postgres.test"),
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
		host    = "postgres"
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

	resource "materialize_source_table_postgres" "test_postgres" {
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

func testAccSourceTablePostgresResource(nameSpace, upstreamName, ownershipRole, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "postgres_password" {
		name  = "%[1]s_secret"
		value = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_postgres" "postgres_connection" {
		name    = "%[1]s_connection"
		host    = "postgres"
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

	resource "materialize_source_table_postgres" "test" {
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
			"id"
		]

		ownership_role = "%[3]s"
		comment        = "%[4]s"

		depends_on = [materialize_role.test_role]
	}
	`, nameSpace, upstreamName, ownershipRole, comment)
}

func testAccCheckSourceTablePostgresExists(name string) resource.TestCheckFunc {
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

func testAccCheckAllSourceTablePostgresDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_source_table_postgres" {
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

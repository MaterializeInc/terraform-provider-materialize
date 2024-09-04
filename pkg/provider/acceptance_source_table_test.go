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

func TestAccSourceTable_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableBasicResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table.test"),
					resource.TestMatchResourceAttr("materialize_source_table.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_table.test", "name", nameSpace+"_table"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "schema_name", "public"),
					// resource.TestCheckResourceAttr("materialize_source_table.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s_table"`, nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_table.test", "text_columns.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "text_columns.0", "updated_at"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "upstream_name", "table2"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "upstream_schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_table.test", "comment", ""),
				),
			},
			{
				ResourceName:      "materialize_source_table.test",
				ImportState:       true,
				ImportStateVerify: false,
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

func testAccSourceTableBasicResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "postgres_password" {
		name  = "%[1]s_secret"
		value = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_postgres" "postgres_connection" {
		name    = "%[1]s_connection"
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

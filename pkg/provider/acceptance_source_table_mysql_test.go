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
					testAccCheckSourceTableExists("materialize_source_table_mysql.test_mysql"),
					resource.TestMatchResourceAttr("materialize_source_table_mysql.test_mysql", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_mysql", "name", nameSpace+"_table_mysql"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_mysql", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_mysql", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_mysql", "upstream_name", "mysql_table1"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_mysql", "upstream_schema_name", "shop"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_mysql", "exclude_columns.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_mysql", "exclude_columns.0", "banned"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_mysql", "source.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_mysql", "source.0.name", nameSpace+"_source_mysql"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_mysql", "source.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_mysql", "source.0.database_name", "materialize"),
				),
			},
			{
				ResourceName:      "materialize_source_table_mysql.test_mysql",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSourceTableMySQL_update(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableMySQLResource(nameSpace, "mysql_table2", "mz_system", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table_mysql.test"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "name", nameSpace+"_table"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "upstream_name", "mysql_table2"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "comment", ""),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "source.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "source.0.name", nameSpace+"_source_mysql"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "source.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "source.0.database_name", "materialize"),
				),
			},
			{
				Config: testAccSourceTableMySQLResource(nameSpace, "mysql_table1", nameSpace+"_role", "Updated comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table_mysql.test"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "name", nameSpace+"_table"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "upstream_name", "mysql_table1"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "ownership_role", nameSpace+"_role"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "comment", "Updated comment"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "source.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "source.0.name", nameSpace+"_source_mysql"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test", "source.0.schema_name", "public"),
				),
			},
		},
	})
}

func TestAccSourceTableMySQL_disappears(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceTableDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableMySQLResource(nameSpace, "mysql_table2", "mz_system", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table_mysql.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: materialize.Table,
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

func testAccSourceTableMySQLBasicResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "mysql_password" {
		name  = "%[1]s_secret_mysql"
		value = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_mysql" "mysql_connection" {
		name    = "%[1]s_connection_mysql"
		host    = "mysql"
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

	resource "materialize_source_table_mysql" "test_mysql" {
		name           = "%[1]s_table_mysql"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name = materialize_source_mysql.test_source_mysql.name
		}

		upstream_name         = "mysql_table1"
		upstream_schema_name  = "shop"
		exclude_columns        = ["banned"]
	}
	`, nameSpace)
}

func testAccSourceTableMySQLResource(nameSpace, upstreamName, ownershipRole, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "mysql_password" {
		name  = "%[1]s_secret_mysql"
		value = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_mysql" "mysql_connection" {
		name    = "%[1]s_connection_mysql"
		host    = "mysql"
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

	resource "materialize_role" "test_role" {
		name = "%[1]s_role"
	}

	resource "materialize_source_table_mysql" "test" {
		name           = "%[1]s_table"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name          = materialize_source_mysql.test_source_mysql.name
			schema_name   = "public"
			database_name = "materialize"
		}

		upstream_name         = "%[2]s"
		upstream_schema_name  = "shop"

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
		if r.Type != "materialize_source_table_mysql" {
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

func TestAccSourceTableMySQL_withNumericTypes(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableMySQLWithNumericTypesResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table_mysql.test_numeric"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_numeric", "name", nameSpace+"_table_numeric"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_numeric", "upstream_name", "mysql_table9"),
				),
			},
		},
	})
}

func TestAccSourceTableMySQL_withDateTimeTypes(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableMySQLWithDateTimeTypesResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableExists("materialize_source_table_mysql.test_datetime"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_datetime", "name", nameSpace+"_table_datetime"),
					resource.TestCheckResourceAttr("materialize_source_table_mysql.test_datetime", "upstream_name", "mysql_table10"),
				),
			},
		},
	})
}

func testAccSourceTableMySQLWithNumericTypesResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "mysql_password" {
		name  = "%[1]s_secret"
		value = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_mysql" "mysql_connection" {
		name    = "%[1]s_connection"
		host    = "mysql"
		port    = 3306
		user {
			text = "repluser"
		}
		password {
			name = materialize_secret.mysql_password.name
		}
	}

	resource "materialize_source_mysql" "test_source" {
		name         = "%[1]s_source"
		cluster_name = "quickstart"

		mysql_connection {
			name = materialize_connection_mysql.mysql_connection.name
		}

		table {
			upstream_name        = "mysql_table9"
			upstream_schema_name = "shop"
			name                 = "mysql_table9_local"
		}
	}

	resource "materialize_source_table_mysql" "test_numeric" {
		name           = "%[1]s_table_numeric"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name = materialize_source_mysql.test_source.name
		}

		upstream_name         = "mysql_table9"
		upstream_schema_name  = "shop"
	}
	`, nameSpace)
}

func testAccSourceTableMySQLWithDateTimeTypesResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "mysql_password" {
		name  = "%[1]s_secret"
		value = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_mysql" "mysql_connection" {
		name    = "%[1]s_connection"
		host    = "mysql"
		port    = 3306
		user {
			text = "repluser"
		}
		password {
			name = materialize_secret.mysql_password.name
		}
	}

	resource "materialize_source_mysql" "test_source" {
		name         = "%[1]s_source"
		cluster_name = "quickstart"

		mysql_connection {
			name = materialize_connection_mysql.mysql_connection.name
		}

		table {
			upstream_name        = "mysql_table10"
			upstream_schema_name = "shop"
			name                 = "mysql_table10_local"
		}

		ignore_columns = ["shop.mysql_table10.year_col"]
	}

	resource "materialize_source_table_mysql" "test_datetime" {
		name           = "%[1]s_table_datetime"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name = materialize_source_mysql.test_source.name
		}

		upstream_name         = "mysql_table10"
		upstream_schema_name  = "shop"

		exclude_columns = ["year_col"]
	}
	`, nameSpace)
}

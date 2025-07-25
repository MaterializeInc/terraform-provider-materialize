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

func TestAccSourceMySQL_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceMySQLBasicResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceMySQLExists("materialize_source_mysql.test"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "name", nameSpace+"_source"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%[1]s_source"`, nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.#", "3"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.upstream_name", "mysql_table1"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.upstream_schema_name", "shop"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.name", fmt.Sprintf(`%s_mysql_table1`, nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.upstream_name", "mysql_table2"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.upstream_schema_name", "shop"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.name", fmt.Sprintf(`%s_mysql_table2`, nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.2.upstream_name", "mysql_table4"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.2.upstream_schema_name", "shop"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.2.name", fmt.Sprintf(`%s_mysql_table4`, nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "ignore_columns.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "ignore_columns.0", "shop.mysql_table2.id"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "text_columns.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "text_columns.0", "shop.mysql_table4.status"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "comment", fmt.Sprintf(`%s comment`, nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "cluster_name", "quickstart"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "size", "25cc"),
				),
			},
			{
				ResourceName:      "materialize_source_mysql.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSourceMySQL_disappears(t *testing.T) {
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceMySQLDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceMySQLBasicResource(sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceMySQLExists("materialize_source_mysql.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "SOURCE",
							Name:       sourceName + "_source",
						},
					),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSourceMySQL_update(t *testing.T) {
	initialName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updatedName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceMySQLUpdateResource(initialName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceMySQLExists("materialize_source_mysql.test"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "name", initialName+"_source"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.upstream_name", "mysql_table1"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.upstream_schema_name", "shop"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.name", fmt.Sprintf(`%s_mysql_table1`, initialName)),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.upstream_name", "mysql_table2"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.upstream_schema_name", "shop"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.name", fmt.Sprintf(`%s_mysql_table2`, initialName)),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "comment", fmt.Sprintf(`%s comment`, initialName)),
				),
			},
			{
				Config: testAccSourceMySQLUpdateResource(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceMySQLExists("materialize_source_mysql.test"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "name", updatedName+"_source"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.upstream_name", "mysql_table1"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.upstream_schema_name", "shop"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.name", fmt.Sprintf(`%s_mysql_table1`, updatedName)),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.upstream_name", "mysql_table2"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.upstream_schema_name", "shop"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.name", fmt.Sprintf(`%s_mysql_table2`, updatedName)),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "comment", fmt.Sprintf(`%s comment`, updatedName)),
				),
			},
		},
	})
}

func TestAccSourceMySQL_updateSubsources(t *testing.T) {
	initialName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updatedName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceMySQLInitialSubsources(initialName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceMySQLExists("materialize_source_mysql.test"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "name", initialName+"_source"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.#", "2"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.upstream_name", "mysql_table1"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.upstream_name", "mysql_table2")),
			},
			{
				Config: testAccSourceMySQLUpdatedSubsources(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceMySQLExists("materialize_source_mysql.test"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "name", updatedName+"_source"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.#", "2"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.upstream_name", "mysql_table1"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.upstream_name", "mysql_table3"),
				),
			},
		},
	})
}

func testAccSourceMySQLBasicResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "mysql_password" {
		name          = "%[1]s_secret"
		value         = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_mysql" "test" {
		name = "%[1]s_connection"
		host = "mysql"
		port = 3306
		user {
			text = "repluser"
		}
		password {
			name          = materialize_secret.mysql_password.name
			schema_name   = materialize_secret.mysql_password.schema_name
			database_name = materialize_secret.mysql_password.database_name
		}
		comment  = "object comment"
	}

	resource "materialize_source_mysql" "test" {
		name = "%[1]s_source"
		cluster_name = "quickstart"

		comment = "%[1]s comment"

		mysql_connection {
			name = materialize_connection_mysql.test.name
		}

		ignore_columns = ["shop.mysql_table2.id"]
		text_columns   = ["shop.mysql_table4.status"]

		table {
			upstream_name  		= "mysql_table1"
			upstream_schema_name = "shop"
			name 		= "%[1]s_mysql_table1"
		}
		table {
			upstream_name  		= "mysql_table2"
			upstream_schema_name = "shop"
			name 		= "%[1]s_mysql_table2"
		}
		table {
			upstream_name        = "mysql_table4"
			upstream_schema_name = "shop"
			name                 = "%[1]s_mysql_table4"
		}
	}
	`, nameSpace)
}

func testAccSourceMySQLUpdateResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "mysql_password" {
		name          = "%[1]s_secret"
		value         = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_mysql" "test" {
		name = "%[1]s_connection"
		host = "mysql"
		port = 3306
		user {
			text = "repluser"
		}
		password {
			name          = materialize_secret.mysql_password.name
			schema_name   = materialize_secret.mysql_password.schema_name
			database_name = materialize_secret.mysql_password.database_name
		}
		comment  = "object comment"
	}

	resource "materialize_source_mysql" "test" {
		name = "%[1]s_source"
		cluster_name = "quickstart"

		comment = "%[1]s comment"

		mysql_connection {
			name = materialize_connection_mysql.test.name
		}

		ignore_columns = ["shop.mysql_table2.id"]

		table {
			upstream_name  		= "mysql_table1"
			upstream_schema_name = "shop"
			name 		= "%[1]s_mysql_table1"
		}
		table {
			upstream_name  		= "mysql_table2"
			upstream_schema_name = "shop"
			name 		= "%[1]s_mysql_table2"
		}
	}
	`, nameSpace)
}

func testAccSourceMySQLInitialSubsources(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "mysql_password" {
		name          = "%[1]s_secret"
		value         = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_mysql" "test" {
		name = "%[1]s_connection"
		host = "mysql"
		port = 3306
		user {
			text = "repluser"
		}
		password {
			name          = materialize_secret.mysql_password.name
			schema_name   = materialize_secret.mysql_password.schema_name
			database_name = materialize_secret.mysql_password.database_name
		}
		comment  = "object comment"
	}

	resource "materialize_source_mysql" "test" {
		name = "%[1]s_source"
		cluster_name = "quickstart"

		mysql_connection {
			name = materialize_connection_mysql.test.name
		}

		table {
			upstream_name  		= "mysql_table1"
			upstream_schema_name = "shop"
			name 		= "%[1]s_mysql_table1"
		}
		table {
			upstream_name  		= "mysql_table2"
			upstream_schema_name = "shop"
			name 		= "%[1]s_mysql_table2"
		}
	}
	`, nameSpace)
}

func testAccSourceMySQLUpdatedSubsources(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "mysql_password" {
		name          = "%[1]s_secret"
		value         = "c2VjcmV0Cg=="
	}

	resource "materialize_connection_mysql" "test" {
		name = "%[1]s_connection"
		host = "mysql"
		port = 3306
		user {
			text = "repluser"
		}
		password {
			name          = materialize_secret.mysql_password.name
			schema_name   = materialize_secret.mysql_password.schema_name
			database_name = materialize_secret.mysql_password.database_name
		}
		comment  = "object comment"
	}

	resource "materialize_source_mysql" "test" {
		name = "%[1]s_source"
		cluster_name = "quickstart"

		mysql_connection {
			name = materialize_connection_mysql.test.name
		}

		table {
			upstream_name  		= "mysql_table1"
			upstream_schema_name = "shop"
			name 		= "%[1]s_mysql_table1"
		}
		table {
			upstream_name  		= "mysql_table3"
			upstream_schema_name = "shop"
			name 		= "%[1]s_mysql_table3"
		}
	}
	`, nameSpace)
}

func testAccCheckSourceMySQLExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("source mysql not found: %s", name)
		}
		_, err = materialize.ScanSource(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllSourceMySQLDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_source_mysql" {
			continue
		}

		_, err := materialize.ScanSource(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("source %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

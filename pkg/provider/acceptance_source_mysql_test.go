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
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "database_name", nameSpace+"_database"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "schema_name", nameSpace+"_schema"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "qualified_sql_name", fmt.Sprintf(`"%[1]s_database"."%[1]s_schema"."%[1]s_source"`, nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.#", "2"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.name", "shop.mysql_table1"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.alias", fmt.Sprintf(`%s_mysql_table1`, nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.name", "shop.mysql_table2"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.alias", fmt.Sprintf(`%s_mysql_table2`, nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "comment", ""),
					// Include additional checks for MySQL-specific attributes if necessary
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

// TODO: Fix this test
// func TestAccSourceMySQL_disappears(t *testing.T) {
// 	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:          func() { testAccPreCheck(t) },
// 		ProviderFactories: testAccProviderFactories,
// 		CheckDestroy:      testAccCheckAllSourceMySQLDestroyed,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccSourceMySQLBasicResource(sourceName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckSourceMySQLExists("materialize_source_mysql.test"),
// 					testAccCheckObjectDisappears(
// 						materialize.MaterializeObject{
// 							ObjectType: "SOURCE",
// 							Name:       sourceName,
// 						},
// 					),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

func TestAccSourceMySQL_update(t *testing.T) {
	initialName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updatedName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceMySQLBasicResource(initialName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceMySQLExists("materialize_source_mysql.test"),
				),
			},
			{
				Config: testAccSourceMySQLBasicResource(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceMySQLExists("materialize_source_mysql.test"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "name", updatedName+"_source"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.name", "shop.mysql_table1"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.0.alias", fmt.Sprintf(`%s_mysql_table1`, updatedName)),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.name", "shop.mysql_table2"),
					resource.TestCheckResourceAttr("materialize_source_mysql.test", "table.1.alias", fmt.Sprintf(`%s_mysql_table2`, updatedName)),
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

	resource "materialize_database" "test" {
		name = "%[1]s_database"
	}

	resource "materialize_schema" "test" {
		name = "%[1]s_schema"
		database_name = materialize_database.test.name
	}

	resource "materialize_source_mysql" "test" {
		name = "%[1]s_source"
		schema_name   = materialize_schema.test.name
		database_name = materialize_database.test.name

		cluster_name = "quickstart"

		mysql_connection {
			name = materialize_connection_mysql.test.name
		}

		table {
			name  = "shop.mysql_table1"
			alias = "%[1]s_mysql_table1"
		}
		table {
			name  = "shop.mysql_table2"
			alias = "%[1]s_mysql_table2"
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

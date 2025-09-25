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

func TestAccSourceSQLServer_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceSQLServerBasicResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceSQLServerExists("materialize_source_sqlserver.test"),
					resource.TestMatchResourceAttr("materialize_source_sqlserver.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "name", fmt.Sprintf("%s_source", nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "schema_name", fmt.Sprintf("%s_schema", nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "database_name", fmt.Sprintf("%s_database", nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "qualified_sql_name", fmt.Sprintf(`"%s_database"."%s_schema"."%s_source"`, nameSpace, nameSpace, nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.#", "2"),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.0.upstream_name", "table1"),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.0.name", fmt.Sprintf(`%s_table1`, nameSpace)),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.1.upstream_name", "table2"),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.1.name", fmt.Sprintf(`%s_table2`, nameSpace)),
				),
			},
		},
	})
}

func TestAccSourceSQLServer_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	sourceName := fmt.Sprintf("old_%s", slug)
	newSourceName := fmt.Sprintf("new_%s", slug)
	source2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceSQLServerResource(roleName, secretName, connName, sourceName, source2Name, "mz_system", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceSQLServerExists("materialize_source_sqlserver.test"),
					testAccCheckSourceSQLServerExists("materialize_source_sqlserver.test_role"),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, sourceName)),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.#", "2"),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.0.upstream_name", "table1"),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.0.name", fmt.Sprintf(`%s_table1`, connName)),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.1.upstream_name", "table2"),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.1.name", fmt.Sprintf(`%s_table2`, connName)),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test_role", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test_role", "comment", "Comment"),
				),
			},
			{
				Config: testAccSourceSQLServerResourceUpdate(roleName, secretName, connName, newSourceName, source2Name, roleName, "New Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceSQLServerExists("materialize_source_sqlserver.test"),
					testAccCheckSourceSQLServerExists("materialize_source_sqlserver.test_role"),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "name", newSourceName),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newSourceName)),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.#", "2"),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.0.upstream_name", "table1"),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.0.name", fmt.Sprintf(`%s_table1`, connName)),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.1.upstream_name", "table2"),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "table.1.name", fmt.Sprintf(`%s_table2`, connName)),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test_role", "comment", "New Comment"),
				),
			},
			{
				Config: testAccSourceSQLServerResource(roleName, secretName, connName, newSourceName, source2Name, roleName, "New Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceSQLServerExists("materialize_source_sqlserver.test"),
					testAccCheckSourceSQLServerExists("materialize_source_sqlserver.test_role"),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "name", newSourceName),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newSourceName)),
				),
			},
		},
	})
}

func TestAccSourceSQLServer_ssl(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceSQLServerSSLResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceSQLServerExists("materialize_source_sqlserver.test_ssl"),
					resource.TestMatchResourceAttr("materialize_source_sqlserver.test_ssl", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_sqlserver.test_ssl", "name", fmt.Sprintf("%s_ssl_source", nameSpace)),
				),
			},
		},
	})
}

func TestAccSourceSQLServer_disappears(t *testing.T) {
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	source2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceSQLServerDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceSQLServerResource(roleName, secretName, connName, sourceName, source2Name, roleName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceSQLServerExists("materialize_source_sqlserver.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "SOURCE",
							Name:       sourceName,
						},
					),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSourceSQLServerBasicResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_database" "test" {
		name = "%[1]s_database"
	}

	resource "materialize_schema" "test" {
		name = "%[1]s_schema"
		database_name = materialize_database.test.name
	}

	resource "materialize_role" "test" {
		name = "%[1]s_role"
	}

	resource "materialize_secret" "sqlserver_password" {
		name  = "%[1]s_secret"
		value = "Password123!"
	}

	resource "materialize_cluster" "test" {
		name = "%[1]s_cluster"
		size = "25cc"
	}

	resource "materialize_connection_sqlserver" "test" {
		name = "%[1]s_conn"
		host = "sqlserver"
		port = 1433
		user {
			text = "sa"
		}
		password {
			name          = materialize_secret.sqlserver_password.name
			schema_name   = materialize_secret.sqlserver_password.schema_name
			database_name = materialize_secret.sqlserver_password.database_name
		}
		database = "testdb"
	}

	resource "materialize_source_sqlserver" "test" {
		name = "%[1]s_source"
		schema_name = materialize_schema.test.name
		database_name = materialize_database.test.name

		sqlserver_connection {
			name = materialize_connection_sqlserver.test.name
			schema_name = materialize_connection_sqlserver.test.schema_name
			database_name = materialize_connection_sqlserver.test.database_name
		}

		cluster_name = materialize_cluster.test.name
		table {
			upstream_name  		 = "table1"
			upstream_schema_name = "dbo"
			name 				 = "%[1]s_table1"
		}
		table {
			upstream_name		= "table2"
			upstream_schema_name = "dbo"
			name				= "%[1]s_table2"
		}
		exclude_columns = ["dbo.table1.about"]
		text_columns    = ["dbo.table2.about"]
	}
	`, nameSpace)
}

func testAccSourceSQLServerSSLResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_database" "test_ssl" {
		name = "%[1]s_ssl_database"
	}

	resource "materialize_schema" "test_ssl" {
		name = "%[1]s_ssl_schema"
		database_name = materialize_database.test_ssl.name
	}

	resource "materialize_secret" "sqlserver_password_ssl" {
		name  = "%[1]s_ssl_secret"
		value = "Password123!"
	}

	resource "materialize_cluster" "test_ssl" {
		name = "%[1]s_ssl_cluster"
		size = "25cc"
	}

	resource "materialize_connection_sqlserver" "test_ssl" {
		name = "%[1]s_ssl_conn"
		host = "sqlserver"
		port = 1433
		user {
			text = "sa"
		}
		password {
			name          = materialize_secret.sqlserver_password_ssl.name
			schema_name   = materialize_secret.sqlserver_password_ssl.schema_name
			database_name = materialize_secret.sqlserver_password_ssl.database_name
		}
		database = "testdb"
		ssl_mode = "require"
	}

	resource "materialize_source_sqlserver" "test_ssl" {
		name = "%[1]s_ssl_source"
		schema_name = materialize_schema.test_ssl.name
		database_name = materialize_database.test_ssl.name

		sqlserver_connection {
			name = materialize_connection_sqlserver.test_ssl.name
			schema_name = materialize_connection_sqlserver.test_ssl.schema_name
			database_name = materialize_connection_sqlserver.test_ssl.database_name
		}

		cluster_name = materialize_cluster.test_ssl.name
	}
	`, nameSpace)
}

func testAccSourceSQLServerResource(roleName, secretName, connName, sourceName, source2Name, sourceOwner, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "test" {
		name = "%[1]s"
	}

	resource "materialize_secret" "sqlserver_password" {
		name  = "%[2]s"
		value = "Password123!"
	}

	resource "materialize_connection_sqlserver" "test" {
		name = "%[3]s"
		host = "sqlserver"
		port = 1433
		user {
			text = "sa"
		}
		password {
			name          = materialize_secret.sqlserver_password.name
			schema_name   = materialize_secret.sqlserver_password.schema_name
			database_name = materialize_secret.sqlserver_password.database_name
		}
		database = "testdb"
	}

	resource "materialize_cluster" "test" {
		name = "%[3]s"
		size = "25cc"
	}

	resource "materialize_source_sqlserver" "test" {
		name = "%[4]s"
		sqlserver_connection {
			name = materialize_connection_sqlserver.test.name
		}

		cluster_name = materialize_cluster.test.name
		table {
			upstream_name  		= "table1"
			upstream_schema_name = "dbo"
			name 		= "%[3]s_table1"
		}
		table {
			upstream_name  		= "table2"
			upstream_schema_name = "dbo"
			name 		= "%[3]s_table2"
		}
		exclude_columns = ["dbo.table1.about"]
		text_columns    = ["dbo.table2.about"]
	}

	resource "materialize_source_sqlserver" "test_role" {
		name = "%[5]s"
		sqlserver_connection {
			name = materialize_connection_sqlserver.test.name
		}

		cluster_name = materialize_cluster.test.name
		table {
			upstream_name  		= "table1"
			upstream_schema_name = "dbo"
			name 		= "%[3]s_table_role_1"
		}
		table {
			upstream_name  		= "table2"
			upstream_schema_name = "dbo"
			name 		= "%[3]s_table_role_2"
		}
		exclude_columns = ["dbo.table1.about"]
		text_columns    = ["dbo.table2.about"]
		ownership_role = "%[6]s"
		comment = "%[7]s"

		depends_on = [materialize_role.test]
	}
	`, roleName, secretName, connName, sourceName, source2Name, sourceOwner, comment)
}

func testAccSourceSQLServerResourceUpdate(roleName, secretName, connName, sourceName, source2Name, sourceOwner, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "test" {
		name = "%[1]s"
	}

	resource "materialize_secret" "sqlserver_password" {
		name          = "%[2]s"
		value         = "Password123!"
	}

	resource "materialize_cluster" "test" {
		name = "%[3]s"
		size = "25cc"
	}

	resource "materialize_connection_sqlserver" "test" {
		name = "%[3]s"
		host = "sqlserver"
		port = 1433
		user {
			text = "sa"
		}
		password {
			name          = materialize_secret.sqlserver_password.name
			schema_name   = materialize_secret.sqlserver_password.schema_name
			database_name = materialize_secret.sqlserver_password.database_name
		}
		database = "testdb"
	}

	resource "materialize_source_sqlserver" "test" {
		name = "%[4]s"
		sqlserver_connection {
			name = materialize_connection_sqlserver.test.name
		}

		cluster_name = materialize_cluster.test.name
		table {
			upstream_name  		= "table1"
			upstream_schema_name = "dbo"
			name 		= "%[3]s_table1"
		}
		table {
			upstream_name  		= "table2"
			upstream_schema_name = "dbo"
			name 		= "%[3]s_table2"
		}
		exclude_columns = ["dbo.table1.about"]
	}

	resource "materialize_source_sqlserver" "test_role" {
		name = "%[5]s"
		sqlserver_connection {
			name = materialize_connection_sqlserver.test.name
		}

		cluster_name = materialize_cluster.test.name
		table {
			upstream_name  		= "table1"
			upstream_schema_name = "dbo"
			name 		= "%[3]s_table_role_1"
		}
		table {
			upstream_name  		= "table2"
			upstream_schema_name = "dbo"
			name 		= "%[3]s_table_role_2"
		}
		exclude_columns = ["dbo.table1.about"]
		ownership_role = "%[6]s"
		comment = "%[7]s"

		depends_on = [materialize_role.test]
	}
	`, roleName, secretName, connName, sourceName, source2Name, sourceOwner, comment)
}

func testAccSourceSQLServerResourceSchema(sourceName string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "test" {
		name  = "%[1]s_secret"
		value = "Password123!"
	}

	resource "materialize_cluster" "test" {
		name = "%[1]s_cluster"
		size = "25cc"
	}

	resource "materialize_connection_sqlserver" "test" {
		name = "%[1]s_conn"
		host = "sqlserver"
		port = 1433
		user {
			text = "sa"
		}
		password {
			name          = materialize_secret.test.name
			schema_name   = materialize_secret.test.schema_name
			database_name = materialize_secret.test.database_name
		}
		database = "testdb"
	}

	resource "materialize_source_sqlserver" "test" {
		name = "%[1]s_source"
		cluster_name = materialize_cluster.test.name
		sqlserver_connection {
			name          = materialize_connection_sqlserver.test.name
			schema_name   = materialize_connection_sqlserver.test.schema_name
			database_name = materialize_connection_sqlserver.test.database_name
		}
		table {
			upstream_name  		= "table1"
			upstream_schema_name = "dbo"
			name 		= "%[1]s_table1"
		}
		exclude_columns = ["dbo.table1.about"]
	}
	`, sourceName)
}

func testAccCheckSourceSQLServerExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("source sqlserver not found: %s", name)
		}
		_, err = materialize.ScanSource(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllSourceSQLServerDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_source_sqlserver" {
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

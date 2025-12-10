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

func TestAccSourceTableSQLServer_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableSQLServerBasicResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableSQLServerExists("materialize_source_table_sqlserver.test_sqlserver"),
					resource.TestMatchResourceAttr("materialize_source_table_sqlserver.test_sqlserver", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_sqlserver", "name", nameSpace+"_table_sqlserver"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_sqlserver", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_sqlserver", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_sqlserver", "upstream_name", "table1"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_sqlserver", "upstream_schema_name", "dbo"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_sqlserver", "source.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_sqlserver", "source.0.name", nameSpace+"_source_sqlserver"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_sqlserver", "source.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_sqlserver", "source.0.database_name", "materialize"),
				),
			},
			{
				ResourceName:      "materialize_source_table_sqlserver.test_sqlserver",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSourceTableSQLServer_update(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableSQLServerResource(nameSpace, "table1", "mz_system", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableSQLServerExists("materialize_source_table_sqlserver.test"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "name", nameSpace+"_table"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "upstream_name", "table1"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "text_columns.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "comment", ""),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "source.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "source.0.name", nameSpace+"_source"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "source.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "source.0.database_name", "materialize"),
				),
			},
			{
				Config: testAccSourceTableSQLServerResource(nameSpace, "table2", nameSpace+"_role", "Updated comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableSQLServerExists("materialize_source_table_sqlserver.test"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "name", nameSpace+"_table"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "upstream_name", "table2"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "text_columns.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "ownership_role", nameSpace+"_role"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "comment", "Updated comment"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "source.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "source.0.name", nameSpace+"_source"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test", "source.0.schema_name", "public"),
				),
			},
		},
	})
}

func TestAccSourceTableSQLServer_disappears(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceTableSQLServerDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableSQLServerResource(nameSpace, "table1", "mz_system", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableSQLServerExists("materialize_source_table_sqlserver.test"),
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

func testAccSourceTableSQLServerBasicResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "sqlserver_password" {
		name  = "%[1]s_secret_sqlserver"
		value = "Password123!"
	}

	resource "materialize_connection_sqlserver" "sqlserver_connection" {
		name    = "%[1]s_connection_sqlserver"
		host    = "sqlserver"
		port    = 1433
		user {
			text = "sa"
		}
		password {
			name = materialize_secret.sqlserver_password.name
		}
		database = "testdb"
		validate = false
	}

	resource "materialize_source_sqlserver" "test_source_sqlserver" {
		name         = "%[1]s_source_sqlserver"
		cluster_name = "quickstart"

		sqlserver_connection {
			name = materialize_connection_sqlserver.sqlserver_connection.name
		}
		table {
			upstream_name  = "table1"
			upstream_schema_name = "dbo"
		}
		exclude_columns = ["dbo.table1.about"]
	}

	resource "materialize_source_table_sqlserver" "test_sqlserver" {
		name           = "%[1]s_table_sqlserver"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name = materialize_source_sqlserver.test_source_sqlserver.name
		}

		upstream_name         = "table1"
		upstream_schema_name  = "dbo"
		exclude_columns       = ["about"]
	}
	`, nameSpace)
}

func testAccSourceTableSQLServerResource(nameSpace, upstreamName, ownershipRole, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "sqlserver_password" {
		name  = "%[1]s_secret"
		value = "Password123!"
	}

	resource "materialize_connection_sqlserver" "sqlserver_connection" {
		name    = "%[1]s_connection"
		host    = "sqlserver"
		port    = 1433
		user {
			text = "sa"
		}
		password {
			name          = materialize_secret.sqlserver_password.name
			database_name = materialize_secret.sqlserver_password.database_name
			schema_name   = materialize_secret.sqlserver_password.schema_name
		}
		database = "testdb"
		validate = false
	}

	resource "materialize_source_sqlserver" "test_source" {
		name         = "%[1]s_source"
		cluster_name = "quickstart"

		sqlserver_connection {
			name          = materialize_connection_sqlserver.sqlserver_connection.name
			schema_name   = materialize_connection_sqlserver.sqlserver_connection.schema_name
			database_name = materialize_connection_sqlserver.sqlserver_connection.database_name
		}
		table {
			upstream_name  = "%[2]s"
			upstream_schema_name = "dbo"
		}
		exclude_columns = ["dbo.%[2]s.about"]
	}

	resource "materialize_role" "test_role" {
		name = "%[1]s_role"
	}

	resource "materialize_source_table_sqlserver" "test" {
		name           = "%[1]s_table"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name          = materialize_source_sqlserver.test_source.name
			schema_name   = "public"
			database_name = "materialize"
		}

		upstream_name         = "%[2]s"
		upstream_schema_name  = "dbo"

		text_columns = [
			"name"
		]

		exclude_columns = ["about"]

		ownership_role = "%[3]s"
		comment        = "%[4]s"

		depends_on = [materialize_role.test_role]
	}
	`, nameSpace, upstreamName, ownershipRole, comment)
}

func testAccCheckSourceTableSQLServerExists(name string) resource.TestCheckFunc {
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
		_, err = materialize.ScanSourceTableSQLServer(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllSourceTableSQLServerDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_source_table_sqlserver" {
			continue
		}

		_, err := materialize.ScanSourceTableSQLServer(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("source table %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

func TestAccSourceTableSQLServer_withUnsupportedTypes(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableSQLServerWithUnsupportedTypesResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableSQLServerExists("materialize_source_table_sqlserver.test_unsupported"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_unsupported", "name", nameSpace+"_table_unsupported"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_unsupported", "upstream_name", "table5"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_unsupported", "exclude_columns.#", "4"),
				),
			},
		},
	})
}

func TestAccSourceTableSQLServer_withNumericTypes(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableSQLServerWithNumericTypesResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableSQLServerExists("materialize_source_table_sqlserver.test_numeric"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_numeric", "name", nameSpace+"_table_numeric"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_numeric", "upstream_name", "table8"),
				),
			},
		},
	})
}

func TestAccSourceTableSQLServer_withDateTimeTypes(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableSQLServerWithDateTimeTypesResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableSQLServerExists("materialize_source_table_sqlserver.test_datetime"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_datetime", "name", nameSpace+"_table_datetime"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_datetime", "upstream_name", "table9"),
				),
			},
		},
	})
}

func TestAccSourceTableSQLServer_withTextTypes(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableSQLServerWithTextTypesResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableSQLServerExists("materialize_source_table_sqlserver.test_text"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_text", "name", nameSpace+"_table_text"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_text", "upstream_name", "table10"),
					resource.TestCheckResourceAttr("materialize_source_table_sqlserver.test_text", "exclude_columns.#", "2"),
				),
			},
		},
	})
}

func testAccSourceTableSQLServerWithUnsupportedTypesResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "sqlserver_password" {
		name  = "%[1]s_secret"
		value = "Password123!"
	}

	resource "materialize_connection_sqlserver" "sqlserver_connection" {
		name    = "%[1]s_connection"
		host    = "sqlserver"
		port    = 1433
		user {
			text = "sa"
		}
		password {
			name = materialize_secret.sqlserver_password.name
		}
		database = "testdb"
		validate = false
	}

	resource "materialize_source_sqlserver" "test_source" {
		name         = "%[1]s_source"
		cluster_name = "quickstart"

		sqlserver_connection {
			name = materialize_connection_sqlserver.sqlserver_connection.name
		}
		table {
			upstream_name  = "table5"
			upstream_schema_name = "dbo"
		}
		exclude_columns = ["dbo.table5.large_text", "dbo.table5.image_data", "dbo.table5.xml_data", "dbo.table5.json_data"]
	}

	resource "materialize_source_table_sqlserver" "test_unsupported" {
		name           = "%[1]s_table_unsupported"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name = materialize_source_sqlserver.test_source.name
		}

		upstream_name         = "table5"
		upstream_schema_name  = "dbo"

		exclude_columns = [
			"large_text",
			"image_data",
			"xml_data",
			"json_data"
		]
	}
	`, nameSpace)
}

func testAccSourceTableSQLServerWithNumericTypesResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "sqlserver_password" {
		name  = "%[1]s_secret"
		value = "Password123!"
	}

	resource "materialize_connection_sqlserver" "sqlserver_connection" {
		name    = "%[1]s_connection"
		host    = "sqlserver"
		port    = 1433
		user {
			text = "sa"
		}
		password {
			name = materialize_secret.sqlserver_password.name
		}
		database = "testdb"
		validate = false
	}

	resource "materialize_source_sqlserver" "test_source" {
		name         = "%[1]s_source"
		cluster_name = "quickstart"

		sqlserver_connection {
			name = materialize_connection_sqlserver.sqlserver_connection.name
		}
		table {
			upstream_name  = "table8"
			upstream_schema_name = "dbo"
		}
	}

	resource "materialize_source_table_sqlserver" "test_numeric" {
		name           = "%[1]s_table_numeric"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name = materialize_source_sqlserver.test_source.name
		}

		upstream_name         = "table8"
		upstream_schema_name  = "dbo"
	}
	`, nameSpace)
}

func testAccSourceTableSQLServerWithDateTimeTypesResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "sqlserver_password" {
		name  = "%[1]s_secret"
		value = "Password123!"
	}

	resource "materialize_connection_sqlserver" "sqlserver_connection" {
		name    = "%[1]s_connection"
		host    = "sqlserver"
		port    = 1433
		user {
			text = "sa"
		}
		password {
			name = materialize_secret.sqlserver_password.name
		}
		database = "testdb"
		validate = false
	}

	resource "materialize_source_sqlserver" "test_source" {
		name         = "%[1]s_source"
		cluster_name = "quickstart"

		sqlserver_connection {
			name = materialize_connection_sqlserver.sqlserver_connection.name
		}
		table {
			upstream_name  = "table9"
			upstream_schema_name = "dbo"
		}
	}

	resource "materialize_source_table_sqlserver" "test_datetime" {
		name           = "%[1]s_table_datetime"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name = materialize_source_sqlserver.test_source.name
		}

		upstream_name         = "table9"
		upstream_schema_name  = "dbo"
	}
	`, nameSpace)
}

func testAccSourceTableSQLServerWithTextTypesResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "sqlserver_password" {
		name  = "%[1]s_secret"
		value = "Password123!"
	}

	resource "materialize_connection_sqlserver" "sqlserver_connection" {
		name    = "%[1]s_connection"
		host    = "sqlserver"
		port    = 1433
		user {
			text = "sa"
		}
		password {
			name = materialize_secret.sqlserver_password.name
		}
		database = "testdb"
		validate = false
	}

	resource "materialize_source_sqlserver" "test_source" {
		name         = "%[1]s_source"
		cluster_name = "quickstart"

		sqlserver_connection {
			name = materialize_connection_sqlserver.sqlserver_connection.name
		}
		table {
			upstream_name  = "table10"
			upstream_schema_name = "dbo"
		}
		exclude_columns = ["dbo.table10.text_col", "dbo.table10.nvarchar_max"]
	}

	resource "materialize_source_table_sqlserver" "test_text" {
		name           = "%[1]s_table_text"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name = materialize_source_sqlserver.test_source.name
		}

		upstream_name         = "table10"
		upstream_schema_name  = "dbo"

		exclude_columns = [
			"text_col",
			"nvarchar_max"
		]
	}
	`, nameSpace)
}

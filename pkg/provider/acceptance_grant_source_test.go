package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jmoiron/sqlx"
)

func TestAccGrantSource_basic(t *testing.T) {
	privilege := randomPrivilege("SOURCE")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantSourceResource(roleName, sourceName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantSourceExists("materialize_grant_source.source_grant", roleName, sourceName, schemaName, databaseName, privilege),
					resource.TestCheckResourceAttr("materialize_grant_source.source_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_grant_source.source_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_grant_source.source_grant", "source_name", sourceName),
					resource.TestCheckResourceAttr("materialize_grant_source.source_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_grant_source.source_grant", "database_name", databaseName),
				),
			},
		},
	})
}

func TestAccGrantSource_disappears(t *testing.T) {
	privilege := randomPrivilege("SOURCE")
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantSourceResource(roleName, sourceName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantSourceExists("materialize_grant_source.source_grant", roleName, sourceName, schemaName, databaseName, privilege),
					testAccCheckGrantSourceRevoked(roleName, sourceName, schemaName, databaseName, privilege),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGrantSourceResource(roleName, sourceName, schemaName, databaseName, privilege string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%s"
}

resource "materialize_database" "test" {
	name = "%s"
}

resource "materialize_schema" "test" {
	name = "%s"
	database_name = materialize_database.test.name
}

resource "materialize_source_load_generator" "test" {
	name                = "%s"
	schema_name         = materialize_schema.test.name
	database_name       = materialize_database.test.name
	size                = "1"
	load_generator_type = "COUNTER"
  
	counter_options {
	  tick_interval = "500ms"
	}
}

resource "materialize_grant_source" "source_grant" {
	role_name     = materialize_role.test.name
	privilege     = "%s"
	database_name = materialize_database.test.name
	schema_name   = materialize_schema.test.name
	source_name   = materialize_source_load_generator.test.name
}
`, roleName, databaseName, schemaName, sourceName, privilege)
}

func testAccCheckGrantSourceExists(grantName, roleName, sourceName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, ok := s.RootModule().Resources[grantName]
		if !ok {
			return fmt.Errorf("grant not found")
		}

		id, err := materialize.SourceId(db, sourceName, schemaName, databaseName)
		if err != nil {
			return err
		}

		roleId, err := materialize.RoleId(db, roleName)
		if err != nil {
			return err
		}

		g, err := materialize.ScanPrivileges(db, "SOURCE", id)
		if err != nil {
			return err
		}

		privilegeMap := materialize.ParsePrivileges(g)
		if !materialize.HasPrivilege(privilegeMap[roleId], privilege) {
			return fmt.Errorf("source object %s does not include privilege %s", g, privilege)
		}
		return nil
	}
}

func testAccCheckGrantSourceRevoked(roleName, sourceName, schemaName, databaseName, privilege string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`REVOKE %s ON SOURCE "%s"."%s"."%s" FROM "%s";`, privilege, databaseName, schemaName, sourceName, roleName))
		return err
	}
}

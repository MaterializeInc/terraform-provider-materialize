package provider

import (
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					testAccCheckGrantExists(
						materialize.MaterializeObject{
							ObjectType:   "SOURCE",
							Name:         sourceName,
							SchemaName:   schemaName,
							DatabaseName: databaseName,
						}, "materialize_source_grant.source_grant", roleName, privilege),
					resource.TestMatchResourceAttr("materialize_source_grant.source_grant", "id", terraformGrantIdRegex),
					resource.TestCheckResourceAttr("materialize_source_grant.source_grant", "role_name", roleName),
					resource.TestCheckResourceAttr("materialize_source_grant.source_grant", "privilege", privilege),
					resource.TestCheckResourceAttr("materialize_source_grant.source_grant", "source_name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_grant.source_grant", "schema_name", schemaName),
					resource.TestCheckResourceAttr("materialize_source_grant.source_grant", "database_name", databaseName),
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

	o := materialize.MaterializeObject{
		ObjectType:   "SOURCE",
		Name:         sourceName,
		SchemaName:   schemaName,
		DatabaseName: databaseName,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccGrantSourceResource(roleName, sourceName, schemaName, databaseName, privilege),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(o, "materialize_source_grant.source_grant", roleName, privilege),
					testAccCheckGrantRevoked(o, roleName, privilege),
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

resource "materialize_cluster" "test" {
	name = "source_grant_test"
	size = "3xsmall"
}

resource "materialize_source_load_generator" "test" {
	name                = "%s"
	schema_name         = materialize_schema.test.name
	database_name       = materialize_database.test.name
	cluster_name        = materialize_cluster.test.name
	load_generator_type = "COUNTER"

	counter_options {
	  tick_interval = "500ms"
	}
}

resource "materialize_source_grant" "source_grant" {
	role_name     = materialize_role.test.name
	privilege     = "%s"
	database_name = materialize_database.test.name
	schema_name   = materialize_schema.test.name
	source_name   = materialize_source_load_generator.test.name
}
`, roleName, databaseName, schemaName, sourceName, privilege)
}

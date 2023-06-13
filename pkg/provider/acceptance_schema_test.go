package provider

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jmoiron/sqlx"
)

func TestAccSchema_basic(t *testing.T) {
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaResource(schemaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists("materialize_schema.test"),
					resource.TestCheckResourceAttr("materialize_schema.test", "name", schemaName),
				),
			},
		},
	})
}

func TestAccSchema_disappears(t *testing.T) {
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSchemasDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaResource(schemaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists("materialize_schema.test"),
					resource.TestCheckResourceAttr("materialize_schema.test", "name", schemaName),
					testAccCheckSchemaDisappears(schemaName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSchemaResource(name string) string {
	return fmt.Sprintf(`
resource "materialize_schema" "test" {
	name = "%s"
	database_name = "materialize"
}
`, name)
}

func testAccCheckSchemaExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Schema not found: %s", name)
		}
		_, err := materialize.ScanSchema(db, r.Primary.ID)
		return err
	}
}

func testAccCheckSchemaDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP SCHEMA "%s";`, name))
		return err
	}
}

func testAccCheckAllSchemasDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_schema" {
			continue
		}

		_, err := materialize.ScanSchema(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("Schema %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

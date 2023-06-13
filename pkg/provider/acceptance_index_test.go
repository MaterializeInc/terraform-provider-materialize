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

func TestAccIndex_basic(t *testing.T) {
	view := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	indexName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexResource(view, indexName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists("materialize_index.test"),
					resource.TestCheckResourceAttr("materialize_index.test", "name", indexName),
					resource.TestCheckResourceAttr("materialize_index.test", "method", "ARRANGEMENT"),
					// TODO: Disable
					// resource.TestCheckResourceAttr("materialize_index.test", "obj_name", "index_view"),
					// resource.TestCheckResourceAttr("materialize_index.test", "obj.0.schema_name", "public"),
					// resource.TestCheckResourceAttr("materialize_index.test", "obj.0.database_name", "materialize"),
					// resource.TestCheckResourceAttr("materialize_index.test", "col_expr.0.field", "column"),
					// resource.TestCheckResourceAttr("materialize_index.test", "database_name", "materialize"),
					// resource.TestCheckResourceAttr("materialize_index.test", "schema_name", "public"),
					// resource.TestCheckResourceAttr("materialize_index.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, indexName)),
				),
			},
		},
	})
}

func TestAccIndex_disappears(t *testing.T) {
	view := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	indexName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllIndexesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexResource(view, indexName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists("materialize_index.test"),
					testAccCheckIndexDisappears(indexName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccIndexResource(view, index string) string {
	return fmt.Sprintf(`
resource "materialize_materialized_view" "test" {
	name = "%s"

	statement = <<SQL
  SELECT
	  1 AS id
  SQL
  }

resource "materialize_index" "test" {
	name = "%s"

	obj_name {
		name = materialize_materialized_view.test.name
		schema_name = materialize_materialized_view.test.schema_name
		database_name = materialize_materialized_view.test.database_name
	}

	col_expr {
		field = "id"
	}
}
`, view, index)
}

func testAccCheckIndexExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("index not found: %s", name)
		}
		_, err := materialize.ScanIndex(db, r.Primary.ID)
		return err
	}
}

func testAccCheckIndexDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP INDEX "%s";`, name))
		return err
	}
}

func testAccCheckAllIndexesDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_index" {
			continue
		}

		_, err := materialize.ScanIndex(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("index %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

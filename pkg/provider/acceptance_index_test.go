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

func TestAccIndex_basic(t *testing.T) {
	viewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	indexName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexResource(viewName, indexName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists("materialize_index.test"),
					resource.TestMatchResourceAttr("materialize_index.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_index.test", "name", indexName),
					resource.TestCheckResourceAttr("materialize_index.test", "method", "ARRANGEMENT"),
					resource.TestCheckResourceAttr("materialize_index.test", "obj_name.#", "1"),
					resource.TestCheckResourceAttr("materialize_index.test", "obj_name.0.name", viewName),
					resource.TestCheckResourceAttr("materialize_index.test", "obj_name.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_index.test", "obj_name.0.database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_index.test", "col_expr.#", "1"),
					resource.TestCheckResourceAttr("materialize_index.test", "col_expr.0.field", "id"),
					resource.TestCheckResourceAttr("materialize_index.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_index.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_index.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, indexName)),
					resource.TestCheckResourceAttr("materialize_index.test", "comment", ""),
				),
			},
			{
				ResourceName:      "materialize_index.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccIndex_update(t *testing.T) {
	viewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	indexName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	comment := "index comment"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexResource(viewName, indexName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists("materialize_index.test"),
					resource.TestCheckResourceAttr("materialize_index.test", "comment", ""),
				),
			},
			{
				Config: testAccIndexWithComment(viewName, indexName, comment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists("materialize_index.test"),
					resource.TestCheckResourceAttr("materialize_index.test", "comment", comment),
				),
			},
		},
	})
}

func TestAccIndex_disappears(t *testing.T) {
	viewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	indexName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllIndexDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexResource(viewName, indexName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists("materialize_index.test"),
					testAccCheckIndexDisappears(indexName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccIndexResource(viewName, indexName string) string {
	return fmt.Sprintf(`
	resource "materialize_view" "test" {
		name = "%[1]s"

		statement = <<SQL
	SELECT
		1 AS id
	SQL
	}

	resource "materialize_index" "test" {
		name = "%[2]s"
		cluster_name = "quickstart"

		obj_name {
			name = materialize_view.test.name
			schema_name = materialize_view.test.schema_name
			database_name = materialize_view.test.database_name
		}

		col_expr {
			field = "id"
		}
	}
	`, viewName, indexName)
}

func testAccIndexWithComment(viewName, indexName, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_view" "test" {
		name = "%[1]s"

		statement = <<SQL
	SELECT
		1 AS id
	SQL
	}

	resource "materialize_index" "test" {
		name = "%[2]s"
		cluster_name = "quickstart"
		comment = "%[3]s"

		obj_name {
			name = materialize_view.test.name
			schema_name = materialize_view.test.schema_name
			database_name = materialize_view.test.database_name
		}

		col_expr {
			field = "id"
		}
	}
	`, viewName, indexName, comment)
}

func testAccCheckIndexExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("index not found: %s", name)
		}
		_, err = materialize.ScanIndex(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckIndexDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		_, err = db.Exec(fmt.Sprintf(`DROP INDEX "%s" RESTRICT;`, name))
		return err
	}
}

func testAccCheckAllIndexDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_index" {
			continue
		}

		_, err := materialize.ScanIndex(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("index %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

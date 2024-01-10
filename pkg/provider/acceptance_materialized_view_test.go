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

func TestAccMaterializedView_basic(t *testing.T) {
	viewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	view2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccMaterializedViewResource(roleName, viewName, view2Name, roleName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaterializedViewExists("materialize_materialized_view.test"),
					resource.TestMatchResourceAttr("materialize_materialized_view.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "name", viewName),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, viewName)),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "statement", "SELECT 1 AS id"),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "ownership_role", "mz_system"),
					testAccCheckMaterializedViewExists("materialize_materialized_view.test_role"),
					resource.TestCheckResourceAttr("materialize_materialized_view.test_role", "name", view2Name),
					resource.TestCheckResourceAttr("materialize_materialized_view.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_materialized_view.test_role", "comment", "Comment"),
				),
			},
			{
				ResourceName:            "materialize_materialized_view.test",
				ImportState:             true,
				ImportStateVerify:       false,
				ImportStateVerifyIgnore: []string{"statement"},
			},
		},
	})
}

func TestAccMaterializedView_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	viewName := fmt.Sprintf("old_%s", slug)
	newViewName := fmt.Sprintf("new_%s", slug)
	view2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccMaterializedViewResource(roleName, viewName, view2Name, "mz_system", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaterializedViewExists("materialize_materialized_view.test_role"),
					resource.TestCheckResourceAttr("materialize_materialized_view.test_role", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_materialized_view.test_role", "comment", "Comment"),
				),
			},
			{
				Config: testAccMaterializedViewResource(roleName, newViewName, view2Name, roleName, "New Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaterializedViewExists("materialize_materialized_view.test"),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "name", newViewName),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newViewName)),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "statement", "SELECT 1 AS id"),
					testAccCheckMaterializedViewExists("materialize_materialized_view.test_role"),
					resource.TestCheckResourceAttr("materialize_materialized_view.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_materialized_view.test_role", "comment", "New Comment"),
				),
			},
		},
	})
}

func TestAccMaterializedView_disappears(t *testing.T) {
	viewName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	view2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllMaterializedViewsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccMaterializedViewResource(roleName, viewName, view2Name, roleName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaterializedViewExists("materialize_materialized_view.test"),
					resource.TestCheckResourceAttr("materialize_materialized_view.test", "name", viewName),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "MATERIALIZED VIEW",
							Name:       viewName,
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccMaterializedViewResource(roleName, materializeViewName, materializeView2Name, materializeViewOwner, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "test" {
		name = "%[1]s"
	}

	resource "materialize_materialized_view" "test" {
		name = "%[2]s"
		statement = "SELECT 1 AS id"
		cluster_name = "quickstart"
		not_null_assertion = ["id"]
	}

	resource "materialize_materialized_view" "test_role" {
		name = "%[3]s"
		statement = "SELECT 1 AS id"
		cluster_name = "quickstart"
		ownership_role = "%[4]s"
		comment = "%[5]s"

		depends_on = [materialize_role.test]
	}
	`, roleName, materializeViewName, materializeView2Name, materializeViewOwner, comment)
}

func testAccCheckMaterializedViewExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Materialized View not found: %s", name)
		}
		_, err = materialize.ScanMaterializedView(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllMaterializedViewsDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_materialized_view" {
			continue
		}

		_, err := materialize.ScanMaterializedView(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("Materialized View %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

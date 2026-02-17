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

func TestAccSourceTableWebhook_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableWebhookBasicResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableWebhookExists("materialize_source_table_webhook.test_webhook"),
					resource.TestMatchResourceAttr("materialize_source_table_webhook.test_webhook", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_table_webhook.test_webhook", "name", nameSpace+"_table_webhook"),
					resource.TestCheckResourceAttr("materialize_source_table_webhook.test_webhook", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_table_webhook.test_webhook", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_webhook.test_webhook", "body_format", "JSON"),
					resource.TestCheckResourceAttr("materialize_source_table_webhook.test_webhook", "include_headers.0.all", "true"),
				),
			},
			{
				ResourceName:      "materialize_source_table_webhook.test_webhook",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSourceTableWebhook_update(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableWebhookResource(nameSpace, "mz_system", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableWebhookExists("materialize_source_table_webhook.test"),
					resource.TestCheckResourceAttr("materialize_source_table_webhook.test", "name", nameSpace+"_table"),
					resource.TestCheckResourceAttr("materialize_source_table_webhook.test", "body_format", "JSON"),
					resource.TestCheckResourceAttr("materialize_source_table_webhook.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_table_webhook.test", "comment", ""),
				),
			},
			{
				Config: testAccSourceTableWebhookResource(nameSpace, nameSpace+"_role", "Updated comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableWebhookExists("materialize_source_table_webhook.test"),
					resource.TestCheckResourceAttr("materialize_source_table_webhook.test", "name", nameSpace+"_table"),
					resource.TestCheckResourceAttr("materialize_source_table_webhook.test", "ownership_role", nameSpace+"_role"),
					resource.TestCheckResourceAttr("materialize_source_table_webhook.test", "comment", "Updated comment"),
				),
			},
		},
	})
}

func TestAccSourceTableWebhook_disappears(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceTableWebhookDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableWebhookResource(nameSpace, "mz_system", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableWebhookExists("materialize_source_table_webhook.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: materialize.Table,
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

func testAccSourceTableWebhookBasicResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_source_webhook" "test_source_webhook" {
		name         = "%[1]s_source_webhook"
		cluster_name = "quickstart"
		body_format  = "JSON"
		include_headers {
			all = true
		}
	}

	resource "materialize_source_table_webhook" "test_webhook" {
		name           = "%[1]s_table_webhook"
		schema_name    = "public"
		database_name  = "materialize"
		body_format    = "JSON"
		include_headers {
			all = true
		}
	}
	`, nameSpace)
}

func testAccSourceTableWebhookResource(nameSpace, ownershipRole, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "test_role" {
		name = "%[1]s_role"
	}

	resource "materialize_source_table_webhook" "test" {
		name           = "%[1]s_table"
		schema_name    = "public"
		database_name  = "materialize"
		body_format    = "JSON"
		include_headers {
			all = true
		}
		check_options {
			field {
				body = true
			}
			alias = "bytes"
		}
		check_expression = "bytes IS NOT NULL"
		ownership_role   = "%[2]s"
		comment          = "%[3]s"

		depends_on = [materialize_role.test_role]
	}
	`, nameSpace, ownershipRole, comment)
}

func testAccCheckSourceTableWebhookExists(name string) resource.TestCheckFunc {
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
		_, err = materialize.ScanSourceTableWebhook(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllSourceTableWebhookDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_source_table_webhook" {
			continue
		}

		_, err := materialize.ScanSourceTable(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("source table %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

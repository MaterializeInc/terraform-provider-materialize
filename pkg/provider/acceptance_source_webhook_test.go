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

func TestAccSourceWebhook_basic(t *testing.T) {
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceWebhookResource(sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceWebhookExists("materialize_source_webhook.test"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "cluster_name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "body_format", "json"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "include_headers", "false"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "ownership_role", fmt.Sprintf("%s_role", sourceName)),
				),
			},
			{
				ResourceName:      "materialize_source_webhook.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSourceWebhook_disappears(t *testing.T) {
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	// source2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	// roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	// connName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceWebhookDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceWebhookResource(sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceWebhookExists("materialize_source_webhook.test"),
					testAccCheckObjectDisappears(
						materialize.ObjectSchemaStruct{
							ObjectType: "SOURCE",
							Name:       sourceName,
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSourceWebhook_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	sourceName := fmt.Sprintf("old_%s", slug)
	newSourceName := fmt.Sprintf("new_%s", slug)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceWebhookResource(sourceName),
			},
			{
				Config: testAccSourceWebhookResource(newSourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceWebhookExists("materialize_source_webhook.test"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "name", newSourceName),
				),
			},
		},
	})
}

func testAccSourceWebhookResource(sourceName string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s_role"
}

resource "materialize_secret" "basic_auth" {
	name          = "%[1]s_password"
	value         = "c2VjcmV0Cg=="
}

resource "materialize_cluster" "example_cluster" {
	name = "%[1]s"
	size = "1"
	replication_factor = 1
}

resource "materialize_source_webhook" "test" {
	name = "%[1]s"
	cluster_name = materialize_cluster.example_cluster.name
	body_format = "json"
	include_headers = false
	ownership_role = materialize_role.test.name

	check_options {
		field {
			body = true
		}
		alias = "bytes"
	}

	check_options {
		field {
			headers = true
		}
	}

	check_options {
		field {
			secret {
				name = materialize_secret.basic_auth.name
			}
		}
		alias = "BASIC_HOOK_AUTH"
	}

	check_expression = "headers->'authorization' = BASIC_HOOK_AUTH"

	depends_on = [materialize_role.test]
}
`, sourceName)
}

func testAccCheckSourceWebhookExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("source webhook not found: %s", name)
		}
		_, err := materialize.ScanSource(db, r.Primary.ID)
		return err
	}
}

func testAccCheckAllSourceWebhookDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_source_webhook" {
			continue
		}

		_, err := materialize.ScanSource(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("source %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

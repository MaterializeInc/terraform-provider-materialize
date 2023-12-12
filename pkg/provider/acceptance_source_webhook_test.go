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

func TestAccSourceWebhook_basic(t *testing.T) {
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceWebhookResource(roleName, secretName, clusterName, sourceName, "mz_system", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceWebhookExists("materialize_source_webhook.test"),
					resource.TestMatchResourceAttr("materialize_source_webhook.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "cluster_name", clusterName),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "body_format", "json"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "include_headers.0.only.0", "a"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "include_headers.0.only.1", "b"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "include_headers.0.not.0", "c"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "include_headers.0.not.1", "d"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "comment", "Comment"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "size", ""),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_options.0.field.0.body", "true"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_options.0.alias", "bytes"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_options.1.field.0.headers", "true"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_options.2.field.0.secret.0.name", secretName),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_expression", "headers->'authorization' = BASIC_HOOK_AUTH"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "subsource.#", "0"),
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

func TestAccSourceWebhookSegment_basic(t *testing.T) {
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceWebhookSegmentResource(sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceWebhookExists("materialize_source_webhook.test"),
					resource.TestMatchResourceAttr("materialize_source_webhook.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "cluster_name", "segment_cluster"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "body_format", "json"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "include_headers.0.all", "true"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_options.0.field.0.body", "true"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_options.0.bytes", "true"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_options.1.field.0.headers", "true"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_options.2.field.0.secret.0.name", "segment_basic_auth"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_options.2.alias", "secret"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_options.2.bytes", "true"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_expression", "decode(headers->'x-signature', 'hex') = hmac(body, secret, 'sha1')"),
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

func TestAccSourceWebhookRudderstack_basic(t *testing.T) {
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceWebhookRudderstackResource(sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceWebhookExists("materialize_source_webhook.test"),
					resource.TestMatchResourceAttr("materialize_source_webhook.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "cluster_name", "rudderstack_cluster"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "body_format", "json"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_options.0.field.0.body", "true"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_options.0.alias", "request_body"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_options.1.field.0.headers", "true"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_options.2.field.0.secret.0.name", "rudderstack_basic_auth"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "check_expression", "headers->'authorization' = rudderstack_basic_auth"),
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

func TestAccSourceWebhook_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	sourceName := fmt.Sprintf("old_%s", slug)
	newSourceName := fmt.Sprintf("new_%s", slug)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceWebhookResource(roleName, secretName, clusterName, sourceName, "mz_system", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceWebhookExists("materialize_source_webhook.test"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "comment", "Comment"),
				),
			},
			{
				Config: testAccSourceWebhookResource(roleName, secretName, clusterName, newSourceName, roleName, "New Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceWebhookExists("materialize_source_webhook.test"),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "name", newSourceName),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_source_webhook.test", "comment", "New Comment"),
				),
			},
		},
	})
}

func TestAccSourceWebhook_disappears(t *testing.T) {
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	clusterName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceWebhookDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceWebhookResource(roleName, secretName, clusterName, sourceName, "mz_system", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceWebhookExists("materialize_source_webhook.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
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

func testAccSourceWebhookResource(roleName, secretName, clusterName, sourceName, sourceOwner, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "test" {
		name = "%[1]s"
	}

	resource "materialize_secret" "basic_auth" {
		name          = "%[2]s"
		value         = "c2VjcmV0Cg=="
	}

	resource "materialize_cluster" "example_cluster" {
		name = "%[3]s"
		size = "3xsmall"
	}

	resource "materialize_source_webhook" "test" {
		name = "%[4]s"
		cluster_name = materialize_cluster.example_cluster.name
		body_format = "json"
		ownership_role = "%[5]s"
		comment = "%[6]s"

		include_headers {
			only = ["a", "b"]
			not  = ["c", "d"]
		}

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
	`, roleName, secretName, clusterName, sourceName, sourceOwner, comment)
}

func testAccSourceWebhookSegmentResource(sourceName string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "basic_auth" {
		name          = "segment_basic_auth"
		value         = "c2VjcmV0Cg=="
	}

	resource "materialize_cluster" "example_cluster" {
		name = "segment_cluster"
		size = "3xsmall"
	}

	resource "materialize_source_webhook" "test" {
		name = "%[1]s"
		cluster_name = materialize_cluster.example_cluster.name
		body_format = "json"

		include_header {
			header = "event-type"
			alias = "event_type"
		}

		include_headers {
			all = true
		}

		check_options {
			field {
				body = true
			}
			bytes = true
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
			alias = "secret"
			bytes = true
		}

		check_expression = "decode(headers->'x-signature', 'hex') = hmac(body, secret, 'sha1')"
	}
	`, sourceName)
}

func testAccSourceWebhookRudderstackResource(sourceName string) string {
	return fmt.Sprintf(`
	resource "materialize_secret" "basic_auth" {
		name          = "rudderstack_basic_auth"
		value         = "c2VjcmV0Cg=="
	}

	resource "materialize_cluster" "example_cluster" {
		name = "rudderstack_cluster"
		size = "3xsmall"
	}

	resource "materialize_source_webhook" "test" {
		name = "%[1]s"
		cluster_name = materialize_cluster.example_cluster.name
		body_format = "json"

		check_options {
			field {
				body = true
			}
			alias = "request_body"
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
		}

		check_expression = "headers->'authorization' = rudderstack_basic_auth"
	}
	`, sourceName)
}

func testAccCheckSourceWebhookExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("source webhook not found: %s", name)
		}
		_, err = materialize.ScanSource(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllSourceWebhookDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_source_webhook" {
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

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

func TestAccNetworkPolicy_basic(t *testing.T) {
	policyName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllNetworkPoliciesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkPolicyResource(policyName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkPolicyExists("materialize_network_policy.test"),
					resource.TestMatchResourceAttr("materialize_network_policy.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_network_policy.test", "name", policyName),
					resource.TestCheckResourceAttr("materialize_network_policy.test", "comment", "Comment"),
					// Test the rules
					resource.TestCheckResourceAttr("materialize_network_policy.test", "rule.#", "2"),
					resource.TestCheckResourceAttr("materialize_network_policy.test", "rule.0.name", "new_york"),
					resource.TestCheckResourceAttr("materialize_network_policy.test", "rule.0.action", "allow"),
					resource.TestCheckResourceAttr("materialize_network_policy.test", "rule.0.direction", "ingress"),
					resource.TestCheckResourceAttr("materialize_network_policy.test", "rule.0.address", "1.2.3.4/28"),
					resource.TestCheckResourceAttr("materialize_network_policy.test", "rule.1.name", "minnesota"),
					resource.TestCheckResourceAttr("materialize_network_policy.test", "rule.1.action", "allow"),
					resource.TestCheckResourceAttr("materialize_network_policy.test", "rule.1.direction", "ingress"),
					resource.TestCheckResourceAttr("materialize_network_policy.test", "rule.1.address", "2.3.4.5/32"),
				),
			},
			{
				ResourceName:      "materialize_network_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkPolicy_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	policyName := fmt.Sprintf("old_%s", slug)
	newPolicyName := fmt.Sprintf("new_%s", slug)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllNetworkPoliciesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkPolicyResource(policyName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkPolicyExists("materialize_network_policy.test"),
					resource.TestCheckResourceAttr("materialize_network_policy.test", "comment", "Comment"),
				),
			},
			{
				Config: testAccNetworkPolicyResource(newPolicyName, "New Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkPolicyExists("materialize_network_policy.test"),
					resource.TestCheckResourceAttr("materialize_network_policy.test", "name", newPolicyName),
					resource.TestCheckResourceAttr("materialize_network_policy.test", "comment", "New Comment"),
				),
			},
		},
	})
}

func TestAccNetworkPolicy_disappears(t *testing.T) {
	policyName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllNetworkPoliciesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkPolicyResource(policyName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkPolicyExists("materialize_network_policy.test"),
					resource.TestCheckResourceAttr("materialize_network_policy.test", "name", policyName),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "NETWORK POLICY",
							Name:       policyName,
						},
					),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccNetworkPolicyResource(policyName, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_network_policy" "test" {
		name    = "%[1]s"
		comment = "%[2]s"

		rule {
			name      = "new_york"
			action    = "allow"
			direction = "ingress"
			address   = "1.2.3.4/28"
		}

		rule {
			name      = "minnesota"
			action    = "allow"
			direction = "ingress"
			address   = "2.3.4.5/32"
		}

	}
	`, policyName, comment)
}

func testAccCheckNetworkPolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Network Policy not found: %s", name)
		}
		_, err = materialize.ScanNetworkPolicy(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllNetworkPoliciesDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_network_policy" {
			continue
		}

		_, err := materialize.ScanNetworkPolicy(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("Network Policy %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

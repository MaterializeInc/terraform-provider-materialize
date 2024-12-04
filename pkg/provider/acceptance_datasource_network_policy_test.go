package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceNetworkPolicies_basic(t *testing.T) {
	policyName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNetworkPolicies(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.materialize_network_policy.test", "network_policies.#"),
					resource.TestCheckResourceAttr("data.materialize_network_policy.test", "network_policies.1.name", policyName),
					resource.TestCheckResourceAttr("data.materialize_network_policy.test", "network_policies.1.comment", "Network policy for office locations"),
					resource.TestCheckResourceAttr("data.materialize_network_policy.test", "network_policies.1.rules.#", "2"),
					resource.TestCheckResourceAttr("data.materialize_network_policy.test", "network_policies.1.rules.0.name", "new_york"),
					resource.TestCheckResourceAttr("data.materialize_network_policy.test", "network_policies.1.rules.0.action", "allow"),
					resource.TestCheckResourceAttr("data.materialize_network_policy.test", "network_policies.1.rules.0.direction", "ingress"),
					resource.TestCheckResourceAttr("data.materialize_network_policy.test", "network_policies.1.rules.0.address", "1.2.3.4/28"),
				),
			},
		},
	})
}

func testAccDataSourceNetworkPolicies(policyName string) string {
	return fmt.Sprintf(`
	resource "materialize_network_policy" "test" {
		name    = "%[1]s"
		comment = "Network policy for office locations"

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

	data "materialize_network_policy" "test" {
		depends_on = [materialize_network_policy.test]
	}
	`, policyName)
}

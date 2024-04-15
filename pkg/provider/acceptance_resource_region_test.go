package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceRegion_basic(t *testing.T) {
	resourceName := "materialize_region.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRegionConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "region_id", "aws/us-east-1"),
					resource.TestCheckResourceAttrSet(resourceName, "sql_address"),
					resource.TestCheckResourceAttr(resourceName, "sql_address", "materialized:6877"),
					resource.TestCheckResourceAttrSet(resourceName, "http_address"),
					resource.TestCheckResourceAttr(resourceName, "http_address", "materialized:6875"),
					resource.TestCheckResourceAttr(resourceName, "resolvable", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "enabled_at"),
					resource.TestCheckResourceAttr(resourceName, "region_state", "true"),
				),
			},
		},
	})
}

func TestAccResourceRegion_update(t *testing.T) {
	resourceName := "materialize_region.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRegionConfig(),
				Check:  resource.TestCheckResourceAttr(resourceName, "region_id", "aws/us-east-1"),
			},
			{
				Config: testAccResourceRegionConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "region_id", "aws/us-east-1"),
					resource.TestCheckResourceAttrSet(resourceName, "sql_address"),
					resource.TestCheckResourceAttr(resourceName, "sql_address", "materialized:6877"),
					resource.TestCheckResourceAttrSet(resourceName, "http_address"),
					resource.TestCheckResourceAttr(resourceName, "http_address", "materialized:6875"),
					resource.TestCheckResourceAttr(resourceName, "resolvable", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "enabled_at"),
					resource.TestCheckResourceAttr(resourceName, "region_state", "true"),
				),
			},
		},
	})
}

func testAccResourceRegionConfig() string {
	return `
		resource "materialize_region" "test" {
			region_id = "aws/us-east-1"
		}
	`
}

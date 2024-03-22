package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceRegion_basic(t *testing.T) {
	resourceName := "data.materialize_region.test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRegionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceRegionExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "regions.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "regions.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "regions.0.url"),
					resource.TestCheckResourceAttrSet(resourceName, "regions.0.cloud_provider"),
					resource.TestCheckResourceAttrSet(resourceName, "regions.0.host"),
					resource.TestCheckResourceAttr(resourceName, "regions.0.name", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "regions.0.id", "aws/us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "regions.0.url", "http://cloud:3001"),
					resource.TestCheckResourceAttr(resourceName, "regions.0.cloud_provider", "aws"),
					resource.TestCheckResourceAttr(resourceName, "regions.0.host", "materialized:6877"),
				),
			},
		},
	})
}

func testAccDataSourceRegionConfig() string {
	return `
	data "materialize_region" "test" {}
	`
}

func testAccDataSourceRegionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Region Data ID is set")
		}

		return nil
	}
}

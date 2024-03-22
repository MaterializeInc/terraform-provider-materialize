package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSCIM2Configuration_basic(t *testing.T) {
	source := "okta"
	connectionName := fmt.Sprintf("test-conn-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSCIM2ConfigurationConfig(source, connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2ConfigurationExists("materialize_scim_config.scim_config_example", source, connectionName),
					resource.TestCheckResourceAttr("materialize_scim_config.scim_config_example", "source", source),
					resource.TestCheckResourceAttr("materialize_scim_config.scim_config_example", "connection_name", connectionName),
					resource.TestCheckResourceAttrSet("materialize_scim_config.scim_config_example", "tenant_id"),
					resource.TestCheckResourceAttr("materialize_scim_config.scim_config_example", "sync_to_user_management", "true"),
					resource.TestCheckResourceAttrSet("materialize_scim_config.scim_config_example", "token"),
					resource.TestCheckResourceAttrSet("materialize_scim_config.scim_config_example", "provisioning_url"),
					resource.TestCheckResourceAttrSet("materialize_scim_config.scim_config_example", "created_at"),
				),
			},
		},
	})
}

func TestAccSCIM2Configuration_update(t *testing.T) {
	source := "okta"
	initialConnectionName := fmt.Sprintf("test-conn-%d", time.Now().UnixNano())
	updatedConnectionName := fmt.Sprintf("updated-conn-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSCIM2ConfigurationConfig(source, initialConnectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2ConfigurationExists("materialize_scim_config.scim_config_example", source, initialConnectionName),
					resource.TestCheckResourceAttr("materialize_scim_config.scim_config_example", "connection_name", initialConnectionName),
				),
			},
			{
				Config: testAccSCIM2ConfigurationConfig(source, updatedConnectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2ConfigurationExists("materialize_scim_config.scim_config_example", source, updatedConnectionName),
					resource.TestCheckResourceAttr("materialize_scim_config.scim_config_example", "connection_name", updatedConnectionName),
					resource.TestCheckResourceAttrSet("materialize_scim_config.scim_config_example", "tenant_id"),
					resource.TestCheckResourceAttr("materialize_scim_config.scim_config_example", "sync_to_user_management", "true"),
					resource.TestCheckResourceAttrSet("materialize_scim_config.scim_config_example", "token"),
					resource.TestCheckResourceAttrSet("materialize_scim_config.scim_config_example", "provisioning_url"),
					resource.TestCheckResourceAttrSet("materialize_scim_config.scim_config_example", "created_at"),
				),
			},
		},
	})
}

func TestAccSCIM2Configuration_disappears(t *testing.T) {
	source := "okta"
	connectionName := fmt.Sprintf("test-conn-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSCIM2ConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSCIM2ConfigurationConfig(source, connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSCIM2ConfigurationExists("materialize_scim_config.scim_config_example", source, connectionName),
					testAccCheckUserDestroy,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSCIM2ConfigurationConfig(source, connectionName string) string {
	return fmt.Sprintf(`
resource "materialize_scim_config" "scim_config_example" {
  source = "%s"
  connection_name = "%s"
}
`, source, connectionName)
}

func testAccCheckSCIM2ConfigurationExists(resourceName, source, connectionName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SCIM 2.0 Configuration ID is set")
		}

		meta := testAccProvider.Meta()
		providerMeta, _ := utils.GetProviderMeta(meta)
		client := providerMeta.Frontegg

		configID := rs.Primary.ID
		var config frontegg.SCIM2Configuration
		configurations, err := frontegg.FetchSCIM2Configurations(context.Background(), client)
		for _, configuration := range configurations {
			if configuration.ID == configID {
				config = configuration
				break
			}
		}

		if err != nil {
			return fmt.Errorf("Error fetching SCIM 2.0 Configuration with ID [%s]: %s", configID, err)
		}

		if config.ConnectionName != connectionName || config.Source != source {
			return fmt.Errorf("SCIM 2.0 Configuration not found")
		}

		return nil
	}
}

func testAccCheckSCIM2ConfigurationDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta()
	providerMeta, _ := utils.GetProviderMeta(meta)
	client := providerMeta.Frontegg

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "materialize_scim_config" {
			continue
		}

		configID := rs.Primary.ID
		var config frontegg.SCIM2Configuration
		configurations, err := frontegg.FetchSCIM2Configurations(context.Background(), client)
		for _, configuration := range configurations {
			if configuration.ID == configID {
				config = configuration
				break
			}
		}
		if err == nil {
			return fmt.Errorf("SCIM 2.0 Configuration %s still exists", configID)
		}
		if config.ID != "" {
			return fmt.Errorf("SCIM 2.0 Configuration %s still exists", configID)
		}

	}

	return nil
}

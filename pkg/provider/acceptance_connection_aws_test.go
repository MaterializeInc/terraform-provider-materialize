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

func TestAccConnectionAws_basic(t *testing.T) {
	resourceName := "materialize_connection_aws.aws_conn"
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	endpoint := "http://localhost:4566"
	accessKeyText := "test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAwsResource(connectionName, endpoint, accessKeyText),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAwsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint),
					resource.TestCheckResourceAttr(resourceName, "access_key_id.0.text", accessKeyText),
					resource.TestCheckResourceAttr(resourceName, "validate", "false"),
				),
			},
		},
	})
}

func TestAccConnectionAws_update(t *testing.T) {
	resourceName := "materialize_connection_aws.aws_conn"
	initialConnectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updatedConnectionName := initialConnectionName + "_updated"
	endpoint := "http://localhost:4566"
	accessKeyText := "test"
	updatedAccessKeyText := "updated_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAwsResource(initialConnectionName, endpoint, accessKeyText),
				Check:  resource.TestCheckResourceAttr(resourceName, "name", initialConnectionName),
			},
			{
				Config: testAccConnectionAwsResource(updatedConnectionName, endpoint, updatedAccessKeyText),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAwsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedConnectionName),
					resource.TestCheckResourceAttr(resourceName, "access_key_id.0.text", updatedAccessKeyText),
					resource.TestCheckResourceAttr(resourceName, "validate", "false"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint),
				),
			},
		},
	})
}

func testAccConnectionAwsResource(name, endpoint, accessKeyText string) string {
	return fmt.Sprintf(`
resource "materialize_secret" "aws_password" {
  name          = "%[1]s_password"
  value         = "secret"
  database_name = "materialize"
  schema_name   = "public"
}

resource "materialize_connection_aws" "aws_conn" {
  name     = "%[1]s"
  endpoint = "%[2]s"
  access_key_id {
    text = "%[3]s"
  }
  secret_access_key {
    name          = materialize_secret.aws_password.name
    database_name = materialize_secret.aws_password.database_name
    schema_name   = materialize_secret.aws_password.schema_name
  }
  validate = false
}
`, name, endpoint, accessKeyText)
}

func testAccCheckConnectionAwsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Connection AWS ID is set")
		}

		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}

		_, err = materialize.ScanConnectionAws(db, utils.ExtractId(rs.Primary.ID))
		if err != nil {
			return fmt.Errorf("Connection AWS (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckConnectionAwsDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "materialize_connection_aws" {
			continue
		}

		_, err := materialize.ScanConnectionAws(db, utils.ExtractId(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("Connection AWS %s still exists", rs.Primary.ID)
		}

		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

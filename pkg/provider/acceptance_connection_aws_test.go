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
	accessKeyText := "test_access_key"
	region := "us-east-1"
	secretKeyText := "test_secret_key"
	sessionToken := "test_session_token"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAwsResource(connectionName, endpoint, accessKeyText, region, secretKeyText, sessionToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAwsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint),
					resource.TestCheckResourceAttr(resourceName, "access_key_id.0.text", accessKeyText),
					resource.TestCheckResourceAttr(resourceName, "aws_region", region),
					resource.TestCheckResourceAttr(resourceName, "secret_access_key.0.name", connectionName+"_secret_access_key"),
					resource.TestCheckResourceAttr(resourceName, "session_token.0.secret.0.name", connectionName+"_session_token"),
				),
			},
		},
	})
}

func TestAccConnectionAws_update(t *testing.T) {
	resourceName := "materialize_connection_aws.aws_conn"
	initialConnectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updatedConnectionName := initialConnectionName + "_updated"
	initialEndpoint := "http://localhost:4566"
	updatedEndpoint := "http://localhost:4567"
	initialAccessKeyText := "test"
	updatedAccessKeyText := "updated_test"
	initialRegion := "us-east-1"
	updatedRegion := "us-west-2"
	initialSecretKeyText := "secret_key"
	updatedSecretKeyText := "updated_secret_key"
	initialSessionToken := "session_token"
	updatedSessionToken := "updated_session_token"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAwsResource(initialConnectionName, initialEndpoint, initialAccessKeyText, initialRegion, initialSecretKeyText, initialSessionToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAwsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", initialConnectionName),
					resource.TestCheckResourceAttr(resourceName, "endpoint", initialEndpoint),
					resource.TestCheckResourceAttr(resourceName, "access_key_id.0.text", initialAccessKeyText),
					resource.TestCheckResourceAttr(resourceName, "aws_region", initialRegion),
					resource.TestCheckResourceAttr(resourceName, "secret_access_key.0.name", initialConnectionName+"_secret_access_key"),
					resource.TestCheckResourceAttr(resourceName, "session_token.0.secret.0.name", initialConnectionName+"_session_token"),
				),
			},
			{
				Config: testAccConnectionAwsResource(updatedConnectionName, updatedEndpoint, updatedAccessKeyText, updatedRegion, updatedSecretKeyText, updatedSessionToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAwsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedConnectionName),
					resource.TestCheckResourceAttr(resourceName, "endpoint", updatedEndpoint),
					resource.TestCheckResourceAttr(resourceName, "access_key_id.0.text", updatedAccessKeyText),
					resource.TestCheckResourceAttr(resourceName, "aws_region", updatedRegion),
					resource.TestCheckResourceAttr(resourceName, "secret_access_key.0.name", updatedConnectionName+"_secret_access_key"),
					resource.TestCheckResourceAttr(resourceName, "session_token.0.secret.0.name", updatedConnectionName+"_session_token"),
				),
			},
		},
	})
}

func testAccConnectionAwsResource(name, endpoint, accessKeyText, region, secretKeyText, sessionToken string) string {
	return fmt.Sprintf(`
resource "materialize_secret" "aws_secret_access_key" {
  name          = "%[1]s_secret_access_key"
  value         = "%[5]s"
  database_name = "materialize"
  schema_name   = "public"
}

resource "materialize_secret" "aws_session_token" {
  name          = "%[1]s_session_token"
  value         = "%[6]s"
  database_name = "materialize"
  schema_name   = "public"
}

resource "materialize_connection_aws" "aws_conn" {
  name     = "%[1]s"
  endpoint = "%[2]s"
  access_key_id {
    text = "%[3]s"
  }
  aws_region = "%[4]s"
  secret_access_key {
    name          = materialize_secret.aws_secret_access_key.name
    database_name = materialize_secret.aws_secret_access_key.database_name
    schema_name   = materialize_secret.aws_secret_access_key.schema_name
  }
  session_token {
	secret {
		name          = materialize_secret.aws_session_token.name
		database_name = materialize_secret.aws_session_token.database_name
		schema_name   = materialize_secret.aws_session_token.schema_name
	}
  }
  validate = false
}
`, name, endpoint, accessKeyText, region, secretKeyText, sessionToken)
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

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

func TestAccConnectionIcebergCatalog_basic(t *testing.T) {
	resourceName := "materialize_connection_iceberg_catalog.test"
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	catalogType := "s3tablesrest"
	url := "http://minio:9000/iceberg-test"
	warehouse := "arn:aws:s3tables:us-east-1:123456789012:bucket/iceberg-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckConnectionIcebergCatalogDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionIcebergCatalogResource(connectionName, catalogType, url, warehouse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionIcebergCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "catalog_type", catalogType),
					resource.TestCheckResourceAttr(resourceName, "url", url),
					resource.TestCheckResourceAttr(resourceName, "warehouse", warehouse),
				),
			},
		},
	})
}

func TestAccConnectionIcebergCatalog_update(t *testing.T) {
	resourceName := "materialize_connection_iceberg_catalog.test"
	initialConnectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updatedConnectionName := initialConnectionName + "_updated"
	catalogType := "s3tablesrest"
	url := "http://minio:9000/iceberg-test"
	warehouse := "arn:aws:s3tables:us-east-1:123456789012:bucket/iceberg-test"

	// TODO: Only the connection name and aws_connection can be updated in-place via ALTER CONNECTION.
	// Changes to catalog_type, url, and warehouse require resource recreation.
	// Error: "storage error: cannot be altered in the requested way (SQLSTATE XX000)"
	// Once Materialize supports ALTER for these properties, add test steps for in-place updates.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckConnectionIcebergCatalogDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionIcebergCatalogResource(initialConnectionName, catalogType, url, warehouse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionIcebergCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", initialConnectionName),
					resource.TestCheckResourceAttr(resourceName, "url", url),
					resource.TestCheckResourceAttr(resourceName, "warehouse", warehouse),
				),
			},
			{
				Config: testAccConnectionIcebergCatalogResource(updatedConnectionName, catalogType, url, warehouse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionIcebergCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedConnectionName),
					resource.TestCheckResourceAttr(resourceName, "url", url),
					resource.TestCheckResourceAttr(resourceName, "warehouse", warehouse),
				),
			},
		},
	})
}

func TestAccConnectionIcebergCatalog_updateAwsConnection(t *testing.T) {
	resourceName := "materialize_connection_iceberg_catalog.test"
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	catalogType := "s3tablesrest"
	url := "https://s3tables.us-east-1.amazonaws.com/iceberg"
	warehouse := "arn:aws:s3tables:us-east-1:123456789012:bucket/my-bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckConnectionIcebergCatalogDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionIcebergCatalogResource(connectionName, catalogType, url, warehouse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionIcebergCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "aws_connection.0.name", connectionName+"_aws"),
				),
			},
			{
				Config: testAccConnectionIcebergCatalogResourceWithDifferentAwsConnection(connectionName, catalogType, url, warehouse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionIcebergCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "aws_connection.0.name", connectionName+"_aws_updated"),
				),
			},
		},
	})
}

func TestAccConnectionIcebergCatalog_disappears(t *testing.T) {
	resourceName := "materialize_connection_iceberg_catalog.test"
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	catalogType := "s3tablesrest"
	url := "http://minio:9000/iceberg-test"
	warehouse := "arn:aws:s3tables:us-east-1:123456789012:bucket/iceberg-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckConnectionIcebergCatalogDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionIcebergCatalogResource(connectionName, catalogType, url, warehouse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionIcebergCatalogExists(resourceName),
					testAccCheckConnectionIcebergCatalogDisappears(resourceName),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConnectionIcebergCatalogResource(name, catalogType, url, warehouse string) string {
	return fmt.Sprintf(`
resource "materialize_secret" "aws_secret_access_key" {
  name  = "%[1]s_secret"
  value = "minio123"
}

resource "materialize_connection_aws" "test_aws" {
  name       = "%[1]s_aws"
  endpoint   = "http://minio:9000"
  aws_region = "us-east-1"
  access_key_id {
    text = "minio"
  }
  secret_access_key {
    name          = materialize_secret.aws_secret_access_key.name
    database_name = materialize_secret.aws_secret_access_key.database_name
    schema_name   = materialize_secret.aws_secret_access_key.schema_name
  }
  validate = false
}

resource "materialize_connection_iceberg_catalog" "test" {
  name          = "%[1]s"
  catalog_type  = "%[2]s"
  url           = "%[3]s"
  warehouse     = "%[4]s"
  aws_connection {
    name          = materialize_connection_aws.test_aws.name
    database_name = materialize_connection_aws.test_aws.database_name
    schema_name   = materialize_connection_aws.test_aws.schema_name
  }
  validate = false
}
`, name, catalogType, url, warehouse)
}

func testAccConnectionIcebergCatalogResourceWithDifferentAwsConnection(name, catalogType, url, warehouse string) string {
	return fmt.Sprintf(`
resource "materialize_secret" "aws_secret_access_key" {
  name  = "%[1]s_secret"
  value = "test_secret_key"
}

resource "materialize_secret" "aws_secret_access_key_updated" {
  name  = "%[1]s_secret_updated"
  value = "test_secret_key_updated"
}

resource "materialize_connection_aws" "test_aws" {
  name       = "%[1]s_aws"
  endpoint   = "http://localhost:4566"
  aws_region = "us-east-1"
  access_key_id {
    text = "test_access_key"
  }
  secret_access_key {
    name          = materialize_secret.aws_secret_access_key.name
    database_name = materialize_secret.aws_secret_access_key.database_name
    schema_name   = materialize_secret.aws_secret_access_key.schema_name
  }
  validate = false
}

resource "materialize_connection_aws" "test_aws_updated" {
  name       = "%[1]s_aws_updated"
  endpoint   = "http://localhost:4566"
  aws_region = "us-east-1"
  access_key_id {
    text = "test_access_key_updated"
  }
  secret_access_key {
    name          = materialize_secret.aws_secret_access_key_updated.name
    database_name = materialize_secret.aws_secret_access_key_updated.database_name
    schema_name   = materialize_secret.aws_secret_access_key_updated.schema_name
  }
  validate = false
}

resource "materialize_connection_iceberg_catalog" "test" {
  name          = "%[1]s"
  catalog_type  = "%[2]s"
  url           = "%[3]s"
  warehouse     = "%[4]s"
  aws_connection {
    name          = materialize_connection_aws.test_aws_updated.name
    database_name = materialize_connection_aws.test_aws_updated.database_name
    schema_name   = materialize_connection_aws.test_aws_updated.schema_name
  }
  validate = false
}
`, name, catalogType, url, warehouse)
}

func testAccCheckConnectionIcebergCatalogExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Connection Iceberg Catalog ID is set")
		}

		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}

		_, err = materialize.ScanConnection(db, utils.ExtractId(rs.Primary.ID))
		if err != nil {
			return fmt.Errorf("connection Iceberg Catalog (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckConnectionIcebergCatalogDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}

		_, err = db.Exec(fmt.Sprintf(`DROP CONNECTION "%s"."%s"."%s";`,
			rs.Primary.Attributes["database_name"],
			rs.Primary.Attributes["schema_name"],
			rs.Primary.Attributes["name"]))
		if err != nil {
			return fmt.Errorf("error dropping connection: %s", err)
		}

		return nil
	}
}

func testAccCheckConnectionIcebergCatalogDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "materialize_connection_iceberg_catalog" {
			continue
		}

		_, err := materialize.ScanConnection(db, utils.ExtractId(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("connection Iceberg Catalog %s still exists", rs.Primary.ID)
		}

		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

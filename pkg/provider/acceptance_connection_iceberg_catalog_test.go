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
					resource.TestCheckResourceAttr(resourceName, "catalog_type", catalogType),
					resource.TestCheckResourceAttr(resourceName, "url", url),
					resource.TestCheckResourceAttr(resourceName, "warehouse", warehouse),
					resource.TestCheckResourceAttr(resourceName, "aws_connection.0.name", connectionName+"_aws"),
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
	initialUrl := "https://s3tables.us-east-1.amazonaws.com/iceberg"
	updatedUrl := "https://s3tables.us-west-2.amazonaws.com/iceberg"
	initialWarehouse := "arn:aws:s3tables:us-east-1:123456789012:bucket/my-bucket"
	updatedWarehouse := "arn:aws:s3tables:us-west-2:123456789012:bucket/my-bucket-2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckConnectionIcebergCatalogDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionIcebergCatalogResource(initialConnectionName, catalogType, initialUrl, initialWarehouse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionIcebergCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", initialConnectionName),
					resource.TestCheckResourceAttr(resourceName, "url", initialUrl),
					resource.TestCheckResourceAttr(resourceName, "warehouse", initialWarehouse),
				),
			},
			{
				Config: testAccConnectionIcebergCatalogResource(updatedConnectionName, catalogType, updatedUrl, updatedWarehouse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionIcebergCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedConnectionName),
					resource.TestCheckResourceAttr(resourceName, "url", updatedUrl),
					resource.TestCheckResourceAttr(resourceName, "warehouse", updatedWarehouse),
				),
			},
		},
	})
}

func TestAccConnectionIcebergCatalog_disappears(t *testing.T) {
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
resource "materialize_connection_aws" "test_aws" {
  name     = "%[1]s_aws"
  endpoint = "http://localhost:4566"
  aws_region = "us-east-1"
  access_key_id {
    text = "test_access_key"
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

		_, err = materialize.ScanConnectionIcebergCatalog(db, utils.ExtractId(rs.Primary.ID))
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

		_, err := materialize.ScanConnectionIcebergCatalog(db, utils.ExtractId(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("connection Iceberg Catalog %s still exists", rs.Primary.ID)
		}

		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

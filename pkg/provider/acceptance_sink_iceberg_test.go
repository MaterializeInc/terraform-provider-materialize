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

func TestAccSinkIceberg_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSinkIcebergDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSinkIcebergResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkIcebergExists("materialize_sink_iceberg.test"),
					resource.TestCheckResourceAttr("materialize_sink_iceberg.test", "name", nameSpace+"_sink"),
					resource.TestCheckResourceAttr("materialize_sink_iceberg.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_sink_iceberg.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_sink_iceberg.test", "namespace", "my_namespace"),
					resource.TestCheckResourceAttr("materialize_sink_iceberg.test", "table", "my_table"),
					resource.TestCheckResourceAttr("materialize_sink_iceberg.test", "key.0", "id"),
					resource.TestCheckResourceAttr("materialize_sink_iceberg.test", "key_not_enforced", "true"),
					resource.TestCheckResourceAttr("materialize_sink_iceberg.test", "commit_interval", "10s"),
				),
			},
			{
				ResourceName:      "materialize_sink_iceberg.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSinkIceberg_update(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	initialSinkName := nameSpace + "_sink"
	updatedSinkName := nameSpace + "_sink_updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSinkIcebergDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSinkIcebergResourceWithName(nameSpace, initialSinkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkIcebergExists("materialize_sink_iceberg.test"),
					resource.TestCheckResourceAttr("materialize_sink_iceberg.test", "name", initialSinkName),
				),
			},
			{
				Config: testAccSinkIcebergResourceWithName(nameSpace, updatedSinkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkIcebergExists("materialize_sink_iceberg.test"),
					resource.TestCheckResourceAttr("materialize_sink_iceberg.test", "name", updatedSinkName),
				),
			},
		},
	})
}

func TestAccSinkIceberg_disappears(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSinkIcebergDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSinkIcebergResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkIcebergExists("materialize_sink_iceberg.test"),
					testAccCheckSinkIcebergDisappears("materialize_sink_iceberg.test"),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSinkIcebergResource(nameSpace string) string {
	return testAccSinkIcebergResourceWithName(nameSpace, nameSpace+"_sink")
}

func testAccSinkIcebergResourceWithName(nameSpace, sinkName string) string {
	return fmt.Sprintf(`
resource "materialize_secret" "aws_secret" {
  name  = "%[1]s_aws_secret"
  value = "minio123"
}

resource "materialize_connection_aws" "test_aws" {
  name       = "%[1]s_aws_conn"
  endpoint   = "http://minio:9000"
  aws_region = "us-east-1"
  access_key_id {
    text = "minio"
  }
  secret_access_key {
    name          = materialize_secret.aws_secret.name
    database_name = materialize_secret.aws_secret.database_name
    schema_name   = materialize_secret.aws_secret.schema_name
  }
  validate = false
}

resource "materialize_connection_iceberg_catalog" "test_catalog" {
  name         = "%[1]s_iceberg_catalog"
  catalog_type = "s3tablesrest"
  url          = "http://minio:9000/iceberg-test"
  warehouse    = "arn:aws:s3tables:us-east-1:123456789012:bucket/iceberg-test"
  aws_connection {
    name          = materialize_connection_aws.test_aws.name
    database_name = materialize_connection_aws.test_aws.database_name
    schema_name   = materialize_connection_aws.test_aws.schema_name
  }
  validate = false
}

resource "materialize_cluster" "test_cluster" {
  name = "%[1]s_cluster"
  size = "25cc"
}

resource "materialize_table" "test_table" {
  name          = "%[1]s_table"
  database_name = "materialize"
  schema_name   = "public"
  column {
    name = "id"
    type = "int4"
  }
  column {
    name = "value"
    type = "text"
  }
}

resource "materialize_sink_iceberg" "test" {
  name         = "%[2]s"
  cluster_name = materialize_cluster.test_cluster.name

  from {
    name          = materialize_table.test_table.name
    database_name = materialize_table.test_table.database_name
    schema_name   = materialize_table.test_table.schema_name
  }

  iceberg_catalog_connection {
    name          = materialize_connection_iceberg_catalog.test_catalog.name
    database_name = materialize_connection_iceberg_catalog.test_catalog.database_name
    schema_name   = materialize_connection_iceberg_catalog.test_catalog.schema_name
  }

  namespace = "my_namespace"
  table     = "my_table"

  aws_connection {
    name          = materialize_connection_aws.test_aws.name
    database_name = materialize_connection_aws.test_aws.database_name
    schema_name   = materialize_connection_aws.test_aws.schema_name
  }

  key             = ["id"]
  key_not_enforced = true
  commit_interval = "10s"
}
`, nameSpace, sinkName)
}

func testAccCheckSinkIcebergExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Sink Iceberg ID is set")
		}

		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}

		_, err = materialize.ScanSink(db, utils.ExtractId(rs.Primary.ID))
		if err != nil {
			return fmt.Errorf("sink Iceberg (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSinkIcebergDisappears(resourceName string) resource.TestCheckFunc {
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

		_, err = db.Exec(fmt.Sprintf(`DROP SINK "%s"."%s"."%s";`,
			rs.Primary.Attributes["database_name"],
			rs.Primary.Attributes["schema_name"],
			rs.Primary.Attributes["name"]))
		if err != nil {
			return fmt.Errorf("error dropping sink: %s", err)
		}

		return nil
	}
}

func testAccCheckSinkIcebergDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "materialize_sink_iceberg" {
			continue
		}

		_, err := materialize.ScanSink(db, utils.ExtractId(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("sink Iceberg %s still exists", rs.Primary.ID)
		}

		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

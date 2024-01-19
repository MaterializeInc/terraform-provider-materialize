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

func TestAccSourceLoadGeneratorCounter_basic(t *testing.T) {
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	source2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceLoadGeneratorResource(roleName, sourceName, source2Name, "3xsmall", roleName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceLoadGeneratorExists("materialize_source_load_generator.test"),
					resource.TestMatchResourceAttr("materialize_source_load_generator.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, sourceName)),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "cluster_name", roleName+"_cluster"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "load_generator_type", "COUNTER"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "counter_options.0.tick_interval", "1000ms"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "counter_options.0.scale_factor", "0.1"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "counter_options.0.max_cardinality", "8"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "subsources.#", "0"),
					testAccCheckSourceLoadGeneratorExists("materialize_source_load_generator.test_role"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test_role", "name", source2Name),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test_role", "comment", "Comment"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test_role", "subsources.#", "0"),
				),
			},
			{
				ResourceName:      "materialize_source_load_generator.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSourceLoadGeneratorAuction_basic(t *testing.T) {
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceLoadGeneratorAuctionResource(sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceLoadGeneratorExists("materialize_source_load_generator.test"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "schema_name", "auction"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "size", "3xsmall"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."auction"."%s"`, sourceName)),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "load_generator_type", "AUCTION"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "auction_options.0.tick_interval", "1000ms"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "auction_options.0.scale_factor", "0.1"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "subsource.#", "6"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "subsource.0.schema_name", "auction"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "subsource.0.database_name", "materialize"),
				),
			},
			{
				ResourceName:      "materialize_source_load_generator.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSourceLoadGeneratorMarketing_basic(t *testing.T) {
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceLoadGeneratorMarketingResource(sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceLoadGeneratorExists("materialize_source_load_generator.test"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "schema_name", "marketing"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."marketing"."%s"`, sourceName)),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "size", "3xsmall"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "load_generator_type", "MARKETING"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "marketing_options.0.tick_interval", "1000ms"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "marketing_options.0.scale_factor", "0.1"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "subsource.#", "7"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "subsource.0.schema_name", "marketing"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "subsource.0.database_name", "materialize"),
				),
			},
			{
				ResourceName:      "materialize_source_load_generator.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSourceLoadGeneratorTPCH_basic(t *testing.T) {
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceLoadGeneratorTPCHResource(sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceLoadGeneratorExists("materialize_source_load_generator.test"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "name", sourceName),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "schema_name", "tpch"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."tpch"."%s"`, sourceName)),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "size", "3xsmall"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "load_generator_type", "TPCH"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "tpch_options.0.tick_interval", "1000ms"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "tpch_options.0.scale_factor", "0.1"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "subsource.#", "9"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "subsource.0.schema_name", "tpch"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "subsource.0.database_name", "materialize"),
				),
			},
			{
				ResourceName:      "materialize_source_load_generator.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccSourceLoadGenerator_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	sourceName := fmt.Sprintf("old_%s", slug)
	newSourceName := fmt.Sprintf("new_%s", slug)
	source2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceLoadGeneratorResource(roleName, sourceName, source2Name, "3xsmall", "mz_system", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceLoadGeneratorExists("materialize_source_load_generator.test"),
					testAccCheckSourceLoadGeneratorExists("materialize_source_load_generator.test_role"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test_role", "size", "3xsmall"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test_role", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test_role", "comment", "Comment"),
				),
			},
			{
				Config: testAccSourceLoadGeneratorResource(roleName, newSourceName, source2Name, "2xsmall", roleName, "New Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceLoadGeneratorExists("materialize_source_load_generator.test"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "name", newSourceName),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newSourceName)),
					testAccCheckSourceLoadGeneratorExists("materialize_source_load_generator.test_role"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test_role", "size", "2xsmall"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test_role", "comment", "New Comment"),
				),
			},
		},
	})
}

func TestAccSourceLoadGenerator_disappears(t *testing.T) {
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	source2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceLoadGeneratorsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceLoadGeneratorResource(roleName, sourceName, source2Name, "3xsmall", roleName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceLoadGeneratorExists("materialize_source_load_generator.test"),
					testAccCheckSourceLoadGeneratorDisappears(sourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSourceLoadGeneratorResource(roleName, sourceName, source2Name, size, sourceOwner, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "test" {
		name = "%[1]s"
	}

	resource "materialize_cluster" "test" {
		name               = "%[1]s_cluster"
		size               = "%[4]s"
	}

	resource "materialize_source_load_generator" "test" {
		name = "%[2]s"
		schema_name = "public"
		cluster_name = materialize_cluster.test.name
		load_generator_type = "COUNTER"
		counter_options {
			tick_interval   = "1000ms"
			scale_factor    = 0.1
			max_cardinality = 8
		}
	}

	resource "materialize_source_load_generator" "test_role" {
		name = "%[3]s"
		schema_name = "public"
		cluster_name = materialize_cluster.test.name
		load_generator_type = "COUNTER"
		counter_options {
			tick_interval = "1000ms"
			scale_factor  = 0.1
		}
		ownership_role = "%[5]s"
		comment = "%[6]s"

		depends_on = [materialize_role.test]
	}
	`, roleName, sourceName, source2Name, size, sourceOwner, comment)
}

// Using unique schemas to prevent table name collisions
func testAccSourceLoadGeneratorAuctionResource(sourceName string) string {
	return fmt.Sprintf(`
	resource "materialize_schema" "test" {
		name = "auction"
	}

	resource "materialize_cluster" "test" {
		name               = "auction_cluster"
		size               = "3xsmall"
	}

	resource "materialize_source_load_generator" "test" {
		name = "%[1]s"
		schema_name = materialize_schema.test.name
		cluster_name = materialize_cluster.test.name
		load_generator_type = "AUCTION"
		auction_options {
			tick_interval = "1000ms"
			scale_factor  = 0.1
		}
	}
	`, sourceName)
}

func testAccSourceLoadGeneratorMarketingResource(sourceName string) string {
	return fmt.Sprintf(`
	resource "materialize_schema" "test" {
		name = "marketing"
	}

	resource "materialize_cluster" "test" {
		name               = "marketing_cluster"
		size               = "3xsmall"
	}

	resource "materialize_source_load_generator" "test" {
		name = "%[1]s"
		schema_name = materialize_schema.test.name
		cluster_name = materialize_cluster.test.name
		load_generator_type = "MARKETING"
		marketing_options {
			tick_interval = "1000ms"
			scale_factor  = 0.1
		}
	}
	`, sourceName)
}

func testAccSourceLoadGeneratorTPCHResource(sourceName string) string {
	return fmt.Sprintf(`
	resource "materialize_schema" "test" {
		name = "tpch"
	}

	resource "materialize_cluster" "test" {
		name               = "tpch_cluster"
		size               = "3xsmall"
	}

	resource "materialize_source_load_generator" "test" {
		name = "%[1]s"
		schema_name = materialize_schema.test.name
		cluster_name = materialize_cluster.test.name
		load_generator_type = "TPCH"
		tpch_options {
			tick_interval = "1000ms"
			scale_factor  = 0.1
		}
	}
	`, sourceName)
}

func testAccCheckSourceLoadGeneratorExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("SourceLoadGenerator not found: %s", name)
		}
		_, err = materialize.ScanSource(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckSourceLoadGeneratorDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		_, err = db.Exec(fmt.Sprintf(`DROP SOURCE "%s";`, name))
		return err
	}
}

func testAccCheckAllSourceLoadGeneratorsDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_source_load_generator" {
			continue
		}

		_, err := materialize.ScanSource(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("SourceLoadGenerator %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

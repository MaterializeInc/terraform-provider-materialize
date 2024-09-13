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

func TestAccSourceTableLoadGen_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableLoadGenBasicResource(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableLoadGenExists("materialize_source_table_load_generator.test_loadgen"),
					resource.TestMatchResourceAttr("materialize_source_table_load_generator.test_loadgen", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "name", nameSpace+"_table_loadgen2"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "upstream_name", "bids"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "source.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "source.0.name", nameSpace+"_loadgen2"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "source.0.schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "source.0.database_name", "materialize"),
				),
			},
		},
	})
}

func TestAccSourceTableLoadGen_update(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceTableLoadGenDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableLoadGenResource(nameSpace, "bids", "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableLoadGenExists("materialize_source_table_load_generator.test_loadgen"),
					resource.TestMatchResourceAttr("materialize_source_table_load_generator.test_loadgen", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "name", nameSpace+"_table_loadgen"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "upstream_name", "bids"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "source.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "source.0.name", nameSpace+"_loadgen"),
				),
			},
			{
				Config: testAccSourceTableLoadGenResource(nameSpace, "bids", nameSpace+"_role", "Updated comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableLoadGenExists("materialize_source_table_load_generator.test_loadgen"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "name", nameSpace+"_table_loadgen"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "upstream_name", "bids"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "source.#", "1"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "source.0.name", nameSpace+"_loadgen"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "ownership_role", nameSpace+"_role"),
					resource.TestCheckResourceAttr("materialize_source_table_load_generator.test_loadgen", "comment", "Updated comment"),
				),
			},
		},
	})
}

func TestAccSourceTableLoadGen_disappears(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceTableLoadGenDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceTableLoadGenResource(nameSpace, "bids", "mz_system", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceTableLoadGenExists("materialize_source_table_load_generator.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "TABLE",
							Name:       nameSpace + "_table",
						},
					),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSourceTableLoadGenBasicResource(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_source_load_generator" "test_loadgen" {
		name                = "%[1]s_loadgen2"
		load_generator_type = "AUCTION"

		schema_name    = "public"
		database_name  = "materialize"

		auction_options {
			tick_interval = "500ms"
		}
	}

	resource "materialize_source_table_load_generator" "test_loadgen" {
		name           = "%[1]s_table_loadgen2"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name = materialize_source_load_generator.test_loadgen.name
			schema_name = "public"
			database_name = "materialize"
		}

		upstream_name = "bids"
	}
	`, nameSpace)
}

func testAccSourceTableLoadGenResource(nameSpace, upstreamName, ownershipRole, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_source_load_generator" "test_loadgen" {
		name                = "%[1]s_loadgen"
		load_generator_type = "AUCTION"

		schema_name    = "public"
		database_name  = "materialize"

		auction_options {
			tick_interval = "500ms"
		}
	}

	resource "materialize_role" "test_role" {
		name = "%[1]s_role"
	}

	resource "materialize_source_table_load_generator" "test_loadgen" {
		name           = "%[1]s_table_loadgen"
		schema_name    = "public"
		database_name  = "materialize"

		source {
			name          = materialize_source_load_generator.test_loadgen.name
			schema_name   = "public"
			database_name = "materialize"
		}

		upstream_name = "%[2]s"
		ownership_role = "%[3]s"
		comment = "%[4]s"

		depends_on = [materialize_role.test_role]
	}
	`, nameSpace, upstreamName, ownershipRole, comment)
}

func testAccCheckSourceTableLoadGenExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("source table not found: %s", name)
		}
		_, err = materialize.ScanSourceTable(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllSourceTableLoadGenDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_source_table_load_generator" {
			continue
		}

		_, err := materialize.ScanSourceTable(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("source table %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

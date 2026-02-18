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

func TestAccSchema_basic(t *testing.T) {
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schema2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaResource(roleName, schemaName, schema2Name, roleName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists("materialize_schema.test"),
					resource.TestMatchResourceAttr("materialize_schema.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_schema.test", "name", schemaName),
					resource.TestCheckResourceAttr("materialize_schema.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_schema.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."%s"`, schemaName)),
					resource.TestCheckResourceAttr("materialize_schema.test", "ownership_role", "mz_system"),
					testAccCheckSchemaExists("materialize_schema.test_role"),
					resource.TestCheckResourceAttr("materialize_schema.test_role", "name", schema2Name),
					resource.TestCheckResourceAttr("materialize_schema.test_role", "ownership_role", roleName),
				),
			},
			{
				ResourceName:            "materialize_schema.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"identify_by_name"},
			},
		},
	})
}

func TestAccSchema_identifyByName(t *testing.T) {
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSchemasDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaResourceWithNameAsId(schemaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists("materialize_schema.test_name_as_id"),
					resource.TestCheckResourceAttr("materialize_schema.test_name_as_id", "name", schemaName),
					resource.TestCheckResourceAttr("materialize_schema.test_name_as_id", "identify_by_name", "true"),
					resource.TestCheckResourceAttr("materialize_schema.test_name_as_id", "id", "aws/us-east-1:name:materialize|"+schemaName),
					resource.TestCheckResourceAttr("materialize_schema.test_name_as_id", "database_name", "materialize"),
				),
			},
			{
				ResourceName:      "materialize_schema.test_name_as_id",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"identify_by_name",
				},
			},
		},
	})
}

func TestAccSchema_update(t *testing.T) {
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schema2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaResource(roleName, schemaName, schema2Name, "mz_system", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists("materialize_schema.test"),
					testAccCheckSchemaExists("materialize_schema.test"),
					resource.TestCheckResourceAttr("materialize_schema.test_role", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_schema.test_role", "comment", "Comment"),
				),
			},
			{
				Config: testAccSchemaResource(roleName, schemaName, schema2Name, roleName, "New Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists("materialize_schema.test"),
					testAccCheckSchemaExists("materialize_schema.test"),
					resource.TestCheckResourceAttr("materialize_schema.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_schema.test_role", "comment", "New Comment"),
				),
			},
		},
	})
}

func TestAccSchema_disappears(t *testing.T) {
	schemaName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schema2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSchemasDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaResource(roleName, schemaName, schema2Name, roleName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists("materialize_schema.test"),
					resource.TestCheckResourceAttr("materialize_schema.test", "name", schemaName),
					resource.TestCheckResourceAttr("materialize_schema.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_schema.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."%s"`, schemaName)),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "SCHEMA",
							Name:       schemaName,
						},
					),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSchemaResource(roleName, schemaName, schema2Name, schemaOwner, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "test" {
		name = "%[1]s"
	}

	resource "materialize_schema" "test" {
		name = "%[2]s"
		database_name = "materialize"
	}

	resource "materialize_schema" "test_role" {
		name = "%[3]s"
		database_name = "materialize"
		ownership_role = "%[4]s"
		comment = "%[5]s"

		depends_on = [materialize_role.test]
	}
	`, roleName, schemaName, schema2Name, schemaOwner, comment)
}

func testAccSchemaResourceWithNameAsId(schemaName string) string {
	return fmt.Sprintf(`
	resource "materialize_schema" "test_name_as_id" {
		name             = "%[1]s"
		database_name    = "materialize"
		identify_by_name = true
	}
	`, schemaName)
}

func testAccCheckSchemaExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Schema not found: %s", name)
		}
		identifyByName := false
		if r.Primary.Attributes["identify_by_name"] == "true" {
			identifyByName = true
		}
		_, err = materialize.ScanSchema(db, utils.ExtractId(r.Primary.ID), identifyByName)
		return err
	}
}

func testAccCheckAllSchemasDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_schema" {
			continue
		}

		_, err = materialize.ScanSchema(db, utils.ExtractId(r.Primary.ID), false)
		if err == nil {
			return fmt.Errorf("Schema %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

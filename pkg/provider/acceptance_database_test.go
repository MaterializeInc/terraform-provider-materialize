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

func TestAccDatabase_basic(t *testing.T) {
	for _, roleName := range []string{
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha),
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + "@materialize.com",
	} {
		t.Run(fmt.Sprintf("roleName=%s", roleName), func(t *testing.T) {
			databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
			database2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
			resource.ParallelTest(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: testAccProviderFactories,
				CheckDestroy:      nil,
				Steps: []resource.TestStep{
					{
						Config: testAccDatabaseResource(roleName, databaseName, database2Name, roleName, "Comment"),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckDatabaseExists("materialize_database.test"),
							resource.TestMatchResourceAttr("materialize_database.test", "id", terraformObjectIdRegex),
							resource.TestCheckResourceAttr("materialize_database.test", "name", databaseName),
							resource.TestCheckResourceAttr("materialize_database.test", "ownership_role", "mz_system"),
							testAccCheckDatabaseExists("materialize_database.test_role"),
							resource.TestCheckResourceAttr("materialize_database.test_role", "name", database2Name),
							resource.TestCheckResourceAttr("materialize_database.test_role", "ownership_role", roleName),
							resource.TestCheckResourceAttr("materialize_database.test_role", "comment", "Comment"),
						),
					},
					{
						ResourceName:            "materialize_database.test",
						ImportState:             true,
						ImportStateVerify:       true,
						ImportStateVerifyIgnore: []string{"no_public_schema"},
					},
				},
			})
		})
	}
}

func TestAccDatabase_update(t *testing.T) {
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	database2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseResource(roleName, databaseName, database2Name, "mz_system", "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("materialize_database.test"),
					testAccCheckDatabaseExists("materialize_database.test_role"),
					resource.TestCheckResourceAttr("materialize_database.test_role", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_database.test_role", "comment", "Comment"),
				),
			},
			{
				Config: testAccDatabaseResource(roleName, databaseName, database2Name, roleName, "New Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("materialize_database.test"),
					testAccCheckDatabaseExists("materialize_database.test_role"),
					resource.TestCheckResourceAttr("materialize_database.test_role", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_database.test_role", "comment", "New Comment"),
				),
			},
		},
	})
}

func TestAccDatabase_disappears(t *testing.T) {
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	database2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllDatabasesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseResource(roleName, databaseName, database2Name, roleName, "Comment"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("materialize_database.test"),
					testAccCheckDatabaseExists("materialize_database.test_role"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "DATABASE",
							Name:       databaseName,
						},
					),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDatabase_noPublicSchema(t *testing.T) {
	databaseName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllDatabasesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseNoSchemaResource(databaseName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("materialize_database.no_schema_test"),
					resource.TestCheckResourceAttr("materialize_database.no_schema_test", "name", databaseName),
					resource.TestCheckResourceAttr("materialize_database.no_schema_test", "ownership_role", roleName),
					resource.TestCheckResourceAttr("materialize_database.no_schema_test", "no_public_schema", "true"),
					testAccCheckPublicSchemaNotExists("data.materialize_schema.no_schema_test"),
				),
			},
			{
				ResourceName:            "materialize_database.no_schema_test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"no_public_schema"},
			},
		},
	})
}

func testAccDatabaseNoSchemaResource(databaseName, roleName string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "test" {
		name = "%[2]s"
	}
    resource "materialize_database" "no_schema_test" {
        name = "%[1]s"
        no_public_schema = true
        ownership_role = "%[2]s"

		depends_on = [materialize_role.test]
    }
	data "materialize_schema" "no_schema_test" {
		database_name = materialize_database.no_schema_test.name
	}
    `, databaseName, roleName)
}

func testAccDatabaseResource(roleName, databaseName, databse2Name, databaseOwner, comment string) string {
	return fmt.Sprintf(`
	resource "materialize_role" "test" {
		name = "%[1]s"
	}

	resource "materialize_database" "test" {
		name = "%[2]s"
	}

	resource "materialize_database" "test_role" {
		name = "%[3]s"
		ownership_role = "%[4]s"
		comment = "%[5]s"

		depends_on = [materialize_role.test]
	}
	`, roleName, databaseName, databse2Name, databaseOwner, comment)
}

func testAccCheckDatabaseExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("database not found: %s", name)
		}
		_, err = materialize.ScanDatabase(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllDatabasesDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_database" {
			continue
		}

		_, err := materialize.ScanDatabase(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("database %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

func testAccCheckPublicSchemaNotExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		for i := 0; ; i++ {
			key := fmt.Sprintf("schemas.%d.name", i)
			if name, ok := rs.Primary.Attributes[key]; ok {
				if name == "public" {
					return fmt.Errorf("Public schema exists when it should not")
				}
			} else {
				break
			}
		}
		return nil
	}
}

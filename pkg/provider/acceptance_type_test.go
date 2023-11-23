package provider

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jmoiron/sqlx"
)

func TestAccTypeList_basic(t *testing.T) {
	typeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	type2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTypeResource(roleName, typeName, type2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTypeExists("materialize_type.test"),
					resource.TestCheckResourceAttr("materialize_type.test", "name", typeName),
					resource.TestCheckResourceAttr("materialize_type.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_type.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_type.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, typeName)),
					resource.TestCheckResourceAttr("materialize_type.test", "list_properties.0.element_type", "int4"),
					resource.TestCheckResourceAttr("materialize_type.test", "row_properties.#", "0"),
					resource.TestCheckResourceAttr("materialize_type.test", "list_properties.#", "1"),
					resource.TestCheckResourceAttr("materialize_type.test", "map_properties.#", "0"),
					resource.TestCheckResourceAttr("materialize_type.test", "category", "list"),
					resource.TestCheckResourceAttr("materialize_type.test", "ownership_role", "mz_system"),
					testAccCheckTypeExists("materialize_type.test_role"),
					resource.TestCheckResourceAttr("materialize_type.test_role", "name", type2Name),
					resource.TestCheckResourceAttr("materialize_type.test_role", "ownership_role", roleName),
				),
			},
			{
				ResourceName:      "materialize_type.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccTypeRow_basic(t *testing.T) {
	typeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTypeRowResource(typeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTypeExists("materialize_type.test"),
					resource.TestCheckResourceAttr("materialize_type.test", "name", typeName),
					resource.TestCheckResourceAttr("materialize_type.test", "row_properties.0.field_name", "a"),
					resource.TestCheckResourceAttr("materialize_type.test", "row_properties.0.field_type", "int4"),
					resource.TestCheckResourceAttr("materialize_type.test", "row_properties.1.field_name", "b"),
					resource.TestCheckResourceAttr("materialize_type.test", "row_properties.1.field_type", "text"),
				),
			},
			{
				ResourceName:      "materialize_type.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccTypeMap_basic(t *testing.T) {
	typeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTypeMapResource(typeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTypeExists("materialize_type.test"),
					resource.TestCheckResourceAttr("materialize_type.test", "name", typeName),
					resource.TestCheckResourceAttr("materialize_type.test", "map_properties.0.key_type", "text"),
					resource.TestCheckResourceAttr("materialize_type.test", "map_properties.0.value_type", "int4"),
				),
			},
			{
				ResourceName:      "materialize_type.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccType_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	typeName := fmt.Sprintf("old_%s", slug)
	newTypeName := fmt.Sprintf("new_%s", slug)
	type2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTypeResource(roleName, typeName, type2Name, "mz_system"),
			},
			{
				Config: testAccTypeResource(roleName, newTypeName, type2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTypeExists("materialize_type.test"),
					resource.TestCheckResourceAttr("materialize_type.test", "name", newTypeName),
					resource.TestCheckResourceAttr("materialize_type.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_type.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_type.test", "ownership_role", "mz_system"),
					testAccCheckTypeExists("materialize_type.test_role"),
					resource.TestCheckResourceAttr("materialize_type.test_role", "name", type2Name),
					resource.TestCheckResourceAttr("materialize_type.test_role", "ownership_role", roleName),
				),
			},
		},
	})
}

func TestAccType_disappears(t *testing.T) {
	typeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	type2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllTypesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccTypeResource(roleName, typeName, type2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTypeExists("materialize_type.test"),
					resource.TestCheckResourceAttr("materialize_type.test", "name", typeName),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "TYPE",
							Name:       typeName,
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTypeResource(roleName, typeName, type2Name, typeOwner string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_type" "test" {
	name = "%[2]s"
	list_properties {
		element_type = "int4"
	}
}

resource "materialize_type" "test_role" {
	name = "%[3]s"
	list_properties {
		element_type = "int4"
	}
	ownership_role = "%[4]s"

	depends_on = [materialize_role.test]
}
`, roleName, typeName, type2Name, typeOwner)
}

func testAccTypeRowResource(typeName string) string {
	return fmt.Sprintf(`
resource "materialize_type" "test" {
	name = "%[1]s"
	row_properties {
		field_name  = "a"
		field_type = "int4"
	}
	row_properties {
		field_name  = "b"
		field_type = "text"
	}
}
`, typeName)
}

func testAccTypeMapResource(typeName string) string {
	return fmt.Sprintf(`
resource "materialize_type" "test" {
	name = "%[1]s"
	map_properties {
		key_type   = "text"
		value_type = "int4"
	}
}
`, typeName)
}

func testAccCheckTypeExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Type not found: %s", name)
		}
		_, err := materialize.ScanType(db, r.Primary.ID)
		return err
	}
}

func testAccCheckAllTypesDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_type" {
			continue
		}

		_, err := materialize.ScanType(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("Type %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

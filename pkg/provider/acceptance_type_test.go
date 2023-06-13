package provider

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jmoiron/sqlx"
)

func TestAccType_basic(t *testing.T) {
	typeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccTypeResource(typeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTypeExists("materialize_type.test"),
					resource.TestCheckResourceAttr("materialize_type.test", "name", typeName),
				),
			},
		},
	})
}

func TestAccType_disappears(t *testing.T) {
	typeName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllTypesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccTypeResource(typeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTypeExists("materialize_type.test"),
					resource.TestCheckResourceAttr("materialize_type.test", "name", typeName),
					testAccCheckTypeDisappears(typeName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTypeResource(name string) string {
	return fmt.Sprintf(`
resource "materialize_type" "test" {
	name = "%s"
	list_properties {
		element_type = "int4"
	}
}
`, name)
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

func testAccCheckTypeDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP TYPE "%s";`, name))
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

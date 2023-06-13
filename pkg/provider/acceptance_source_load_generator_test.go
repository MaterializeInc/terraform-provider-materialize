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

func TestAccSourceLoadGenerator_basic(t *testing.T) {
	sourceLoadGeneratorName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceLoadGeneratorResource(sourceLoadGeneratorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceLoadGeneratorExists("materialize_source_load_generator.test"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "name", sourceLoadGeneratorName),
				),
			},
		},
	})
}

func TestAccSourceLoadGenerator_disappears(t *testing.T) {
	sourceLoadGeneratorName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSourceLoadGeneratorsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceLoadGeneratorResource(sourceLoadGeneratorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceLoadGeneratorExists("materialize_source_load_generator.test"),
					resource.TestCheckResourceAttr("materialize_source_load_generator.test", "name", sourceLoadGeneratorName),
					testAccCheckSourceLoadGeneratorDisappears(sourceLoadGeneratorName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSourceLoadGeneratorResource(name string) string {
	return fmt.Sprintf(`
resource "materialize_source_load_generator" "test" {
	name = "%s"
	schema_name = "public"
	size = "1"
	load_generator_type = "COUNTER"
	counter_options {
		tick_interval       = "500ms"
		scale_factor        = 0.01
	}
}
`, name)
}

func testAccCheckSourceLoadGeneratorExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("SourceLoadGenerator not found: %s", name)
		}
		_, err := materialize.ScanSource(db, r.Primary.ID)
		return err
	}
}

func testAccCheckSourceLoadGeneratorDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP SOURCE "%s";`, name))
		return err
	}
}

func testAccCheckAllSourceLoadGeneratorsDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_source_load_generator" {
			continue
		}

		_, err := materialize.ScanSource(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("SourceLoadGenerator %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

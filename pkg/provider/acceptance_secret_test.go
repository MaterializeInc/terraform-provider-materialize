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

func TestAccSecret_basic(t *testing.T) {
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretResource(secretName, "sekret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists("materialize_secret.test"),
					resource.TestCheckResourceAttr("materialize_secret.test", "name", secretName),
					resource.TestCheckResourceAttr("materialize_secret.test", "value", "sekret"),
					resource.TestCheckResourceAttr("materialize_secret.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_secret.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_secret.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, secretName)),
				),
			},
		},
	})
}

func TestAccSecret_update(t *testing.T) {
	secretName := "old"
	newSecretName := "new"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretResource(secretName, "sekret"),
			},
			{
				Config: testAccSecretResource(newSecretName, "sek"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists("materialize_secret.test"),
					resource.TestCheckResourceAttr("materialize_secret.test", "name", newSecretName),
					resource.TestCheckResourceAttr("materialize_secret.test", "value", "sek"),
					resource.TestCheckResourceAttr("materialize_secret.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_secret.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_secret.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newSecretName)),
				),
			},
		},
	})
}

func TestAccSecret_disappears(t *testing.T) {
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSecretsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretResource(secretName, "sekret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists("materialize_secret.test"),
					testAccCheckSecretDisappears(secretName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSecretResource(name, secret string) string {
	return fmt.Sprintf(`
resource "materialize_secret" "test" {
	name = "%s"
	value = "%s"
}
`, name, secret)
}

func testAccCheckSecretExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("secret not found: %s", name)
		}
		_, err := materialize.ScanSecret(db, r.Primary.ID)
		return err
	}
}

func testAccCheckSecretDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP SECRET "%s";`, name))
		return err
	}
}

func testAccCheckAllSecretsDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_secret" {
			continue
		}

		_, err := materialize.ScanSecret(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("secret %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

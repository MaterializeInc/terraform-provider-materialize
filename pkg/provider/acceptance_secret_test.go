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
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	secret2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretResource(roleName, secretName, "sekret", secret2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists("materialize_secret.test"),
					resource.TestCheckResourceAttr("materialize_secret.test", "name", secretName),
					resource.TestCheckResourceAttr("materialize_secret.test", "value", "sekret"),
					resource.TestCheckResourceAttr("materialize_secret.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_secret.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_secret.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, secretName)),
					resource.TestCheckResourceAttr("materialize_secret.test", "ownership_role", "mz_system"),
					testAccCheckSecretExists("materialize_secret.test_role"),
					resource.TestCheckResourceAttr("materialize_secret.test_role", "name", secret2Name),
					resource.TestCheckResourceAttr("materialize_secret.test_role", "ownership_role", roleName),
				),
			},
		},
	})
}

func TestAccSecret_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	secretName := fmt.Sprintf("old_%s", slug)
	newSecretName := fmt.Sprintf("new_%s", slug)
	secret2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretResource(roleName, secretName, "sekret", secret2Name, "mz_system"),
			},
			{
				Config: testAccSecretResource(roleName, newSecretName, "sek", secret2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists("materialize_secret.test"),
					resource.TestCheckResourceAttr("materialize_secret.test", "name", newSecretName),
					resource.TestCheckResourceAttr("materialize_secret.test", "value", "sek"),
					resource.TestCheckResourceAttr("materialize_secret.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_secret.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_secret.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newSecretName)),
					testAccCheckSecretExists("materialize_secret.test_role"),
					resource.TestCheckResourceAttr("materialize_secret.test_role", "ownership_role", roleName),
				),
			},
		},
	})
}

func TestAccSecret_disappears(t *testing.T) {
	secretName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	secret2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllSecretsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretResource(roleName, secretName, "sekret", secret2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists("materialize_secret.test"),
					testAccCheckObjectDisappears(
						materialize.ObjectSchemaStruct{
							ObjectType: "SECRET",
							Name:       secretName,
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSecretResource(roleName, secretName, secretValue, secret2Name, secretOwner string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
	name = "%[1]s"
}

resource "materialize_secret" "test" {
	name = "%[2]s"
	value = "%[3]s"
}

resource "materialize_secret" "test_role" {
	name = "%[4]s"
	value = "%[3]s"
	ownership_role = "%[5]s"

	depends_on = [materialize_role.test]
}
`, roleName, secretName, secretValue, secret2Name, secretOwner)
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

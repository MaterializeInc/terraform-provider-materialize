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

func TestAccConnSshTunnel_basic(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnSshTunnelResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnSshTunnelExists("materialize_connection_ssh_tunnel.test"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "host", "ssh_host"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "user", "ssh_user"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "port", "22"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
				),
			},
		},
	})
}

func TestAccConnSshTunnel_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	connectionName := fmt.Sprintf("old_%s", slug)
	newConnectionName := fmt.Sprintf("new_%s", slug)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnSshTunnelResource(connectionName),
			},
			{
				Config: testAccConnSshTunnelResource(newConnectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnSshTunnelExists("materialize_connection_ssh_tunnel.test"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "name", newConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newConnectionName)),
				),
			},
		},
	})
}

func TestAccConnSshTunnel_disappears(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllConnSshTunnelDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnSshTunnelResource(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnSshTunnelExists("materialize_connection_ssh_tunnel.test"),
					testAccCheckConnSshTunnelDisappears(connectionName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConnSshTunnelResource(name string) string {
	return fmt.Sprintf(`
resource "materialize_connection_ssh_tunnel" "test" {
	name        = "%s"
	schema_name = "public"
	host        = "ssh_host"
	user        = "ssh_user"
	port        = 22
}
`, name)
}

func testAccCheckConnSshTunnelExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("connection ssh tunnel not found: %s", name)
		}
		_, err := materialize.ScanConnectionSshTunnel(db, r.Primary.ID)
		return err
	}
}

func testAccCheckConnSshTunnelDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db := testAccProvider.Meta().(*sqlx.DB)
		_, err := db.Exec(fmt.Sprintf(`DROP CONNECTION "%s";`, name))
		return err
	}
}

func testAccCheckAllConnSshTunnelDestroyed(s *terraform.State) error {
	db := testAccProvider.Meta().(*sqlx.DB)

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_connection_ssh_tunnel" {
			continue
		}

		_, err := materialize.ScanConnectionSshTunnel(db, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("connection %v still exists", r.Primary.ID)
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

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

func TestAccConnSshTunnel_basic(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnSshTunnelResource(roleName, connectionName, "ssh_host", "ssh_user", "22", connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnSshTunnelExists("materialize_connection_ssh_tunnel.test"),
					resource.TestMatchResourceAttr("materialize_connection_ssh_tunnel.test", "id", terraformObjectIdRegex),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "name", connectionName),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "host", "ssh_host"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "user", "ssh_user"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "port", "22"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, connectionName)),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "ownership_role", "mz_system"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "comment", "object comment "+connectionName),
					testAccCheckConnKafkaExists("materialize_connection_ssh_tunnel.test_role"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test_role", "name", connection2Name),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test_role", "ownership_role", roleName),
				),
			},
			{
				ResourceName:      "materialize_connection_ssh_tunnel.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccConnSshTunnel_update(t *testing.T) {
	slug := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	connectionName := fmt.Sprintf("old_%s", slug)
	newConnectionName := fmt.Sprintf("new_%s", slug)
	newHostName := "new_ssh_host"
	newPort := "2222"
	newUser := "new_ssh_user"
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConnSshTunnelResource(roleName, connectionName, "ssh_host", "ssh_user", "22", connection2Name, "mz_system"),
			},
			{
				Config: testAccConnSshTunnelResource(roleName, newConnectionName, newHostName, newUser, newPort, connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnSshTunnelExists("materialize_connection_ssh_tunnel.test"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "name", newConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "host", newHostName),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "user", newUser),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "port", newPort),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "database_name", "materialize"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "schema_name", "public"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "comment", "object comment "+newConnectionName),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test", "qualified_sql_name", fmt.Sprintf(`"materialize"."public"."%s"`, newConnectionName)),
					testAccCheckConnKafkaExists("materialize_connection_ssh_tunnel.test_role"),
					resource.TestCheckResourceAttr("materialize_connection_ssh_tunnel.test_role", "ownership_role", roleName),
				),
			},
		},
	})
}

func TestAccConnSshTunnel_disappears(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connection2Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	roleName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAllConnSshTunnelDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccConnSshTunnelResource(roleName, connectionName, "ssh_host", "ssh_user", "22", connection2Name, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnSshTunnelExists("materialize_connection_ssh_tunnel.test"),
					testAccCheckObjectDisappears(
						materialize.MaterializeObject{
							ObjectType: "CONNECTION",
							Name:       connectionName,
						},
					),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConnSshTunnelResource(roleName, connectionName, host, user, port, connection2Name, connectionOwner string) string {
	return fmt.Sprintf(`
resource "materialize_role" "test" {
    name = "%[1]s"
}

resource "materialize_connection_ssh_tunnel" "test" {
    name        = "%[2]s"
    schema_name = "public"
    host        = "%[3]s"
    user        = "%[4]s"
    port        = %[5]s
    comment     = "object comment %[2]s"
	validate    = false
}

resource "materialize_connection_ssh_tunnel" "test_role" {
    name            = "%[6]s"
    schema_name     = "public"
    host            = "%[3]s"
    user            = "%[4]s"
    port            = %[5]s
    ownership_role  = "%[7]s"
	validate    = false
    depends_on = [materialize_role.test]
}
`, roleName, connectionName, host, user, port, connection2Name, connectionOwner)
}

func testAccCheckConnSshTunnelExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := testAccProvider.Meta()
		db, _, err := utils.GetDBClientFromMeta(meta, nil)
		if err != nil {
			return fmt.Errorf("error getting DB client: %s", err)
		}
		r, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("connection ssh tunnel not found: %s", name)
		}
		_, err = materialize.ScanConnectionSshTunnel(db, utils.ExtractId(r.Primary.ID))
		return err
	}
}

func testAccCheckAllConnSshTunnelDestroyed(s *terraform.State) error {
	meta := testAccProvider.Meta()
	db, _, err := utils.GetDBClientFromMeta(meta, nil)
	if err != nil {
		return fmt.Errorf("error getting DB client: %s", err)
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "materialize_connection_ssh_tunnel" {
			continue
		}

		_, err := materialize.ScanConnectionSshTunnel(db, utils.ExtractId(r.Primary.ID))
		if err == nil {
			return fmt.Errorf("connection %v still exists", utils.ExtractId(r.Primary.ID))
		} else if err != sql.ErrNoRows {
			return err
		}
	}

	return nil
}

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceMaterializedView_basic(t *testing.T) {
	nameSpace := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceMaterializedView(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.materialize_materialized_view.test_database", "database_name", nameSpace),
					resource.TestCheckNoResourceAttr("data.materialize_materialized_view.test_database", "schema_name"),
					resource.TestCheckResourceAttr("data.materialize_materialized_view.test_database", "materialized_views.#", "3"),
					resource.TestCheckResourceAttr("data.materialize_materialized_view.test_database_schema", "database_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_materialized_view.test_database_schema", "schema_name", nameSpace),
					resource.TestCheckResourceAttr("data.materialize_materialized_view.test_database_schema", "materialized_views.#", "2"),
					resource.TestCheckResourceAttr("data.materialize_materialized_view.test_database_2", "database_name", nameSpace+"_2"),
					resource.TestCheckNoResourceAttr("data.materialize_materialized_view.test_database_2", "schema_name"),
					resource.TestCheckResourceAttr("data.materialize_materialized_view.test_database_2", "materialized_views.#", "2"),

					resource.TestCheckNoResourceAttr("data.materialize_materialized_view.test_all", "database_name"),
					resource.TestCheckNoResourceAttr("data.materialize_materialized_view.test_all", "schema_name"),
					// Cannot ensure the number of all objects
				),
			},
		},
	})
}

func testAccDatasourceMaterializedView(nameSpace string) string {
	return fmt.Sprintf(`
	resource "materialize_database" "test" {
		name    = "%[1]s"
	}

	resource "materialize_database" "test_2" {
		name    = "%[1]s_2"
	}

	resource "materialize_schema" "test" {
		name          = "%[1]s"
		database_name = materialize_database.test.name
	}

	resource "materialize_materialized_view" "a" {
		name          = "a"
		database_name = materialize_database.test.name
		cluster_name  = "default"
	  
		statement = <<SQL
	  SELECT
		  1 AS id, 1 AS id_2
	  SQL
	}

	resource "materialize_materialized_view" "b" {
		name          = "b"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		cluster_name  = "default"
	  
		statement = <<SQL
	  SELECT
		  1 AS id
	  SQL
	}

	resource "materialize_materialized_view" "c" {
		name          = "c"
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		cluster_name  = "default"
	  
		statement = <<SQL
	  SELECT
		  1 AS id, 1 AS id_2, 1 AS id_3, 1 AS id_4, 1 AS id_5
	  SQL
	}

	resource "materialize_materialized_view" "d" {
		name          = "d"
		database_name = materialize_database.test_2.name
		cluster_name  = "default"
	  
		statement = <<SQL
	  SELECT
		  1 AS id, 1 AS id_2
	  SQL
	}

	resource "materialize_materialized_view" "e" {
		name          = "e"
		database_name = materialize_database.test_2.name
		cluster_name  = "default"
	  
		statement = <<SQL
	  SELECT
		  1 AS id, 1 AS id_2
	  SQL
	}

	data "materialize_materialized_view" "test_all" {
		depends_on    = [
			materialize_materialized_view.a,
			materialize_materialized_view.b,
			materialize_materialized_view.c,
			materialize_materialized_view.d,
			materialize_materialized_view.e,
		]
	}

	data "materialize_materialized_view" "test_database" {
		database_name = materialize_database.test.name
		depends_on    = [
			materialize_materialized_view.a,
			materialize_materialized_view.b,
			materialize_materialized_view.c,
			materialize_materialized_view.d,
			materialize_materialized_view.e,
		]
	}
	
	data "materialize_materialized_view" "test_database_schema" {
		database_name = materialize_database.test.name
		schema_name   = materialize_schema.test.name
		depends_on    = [
			materialize_materialized_view.a,
			materialize_materialized_view.b,
			materialize_materialized_view.c,
			materialize_materialized_view.d,
			materialize_materialized_view.e,
		]
	}

	data "materialize_materialized_view" "test_database_2" {
		database_name = materialize_database.test_2.name
		depends_on = [
			materialize_materialized_view.a,
			materialize_materialized_view.b,
			materialize_materialized_view.c,
			materialize_materialized_view.d,
			materialize_materialized_view.e,
		]
	}
	`, nameSpace)
}

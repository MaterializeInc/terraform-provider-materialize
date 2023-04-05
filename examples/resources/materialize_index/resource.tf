resource "materialize_index" "loadgen_index" {
  name         = "example_index"
  cluster_name = "cluster"
  method       = "ARRANGEMENT"

  obj_name {
    name          = "source"
    schema_name   = "schema"
    database_name = "database"
  }
}

# CREATE INDEX index
#     IN CLUSTER cluster
#     ON "database"."schema"."source"
#     USING ARRANGEMENT
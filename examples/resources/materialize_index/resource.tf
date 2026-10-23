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

# A materialized view that reads from the index above. Materialize does not
# drop an index that is still read from, so the view has to be dropped first on
# destroy and replaced with the index. Referencing the id, rather than the whole
# resource, avoids replacing the view on in-place changes such as a comment.
resource "materialize_materialized_view" "reads_index" {
  name         = "reads_index"
  cluster_name = "cluster"

  statement = <<SQL
SELECT * FROM "database"."schema"."source"
SQL

  lifecycle {
    replace_triggered_by = [materialize_index.loadgen_index.id]
  }
}

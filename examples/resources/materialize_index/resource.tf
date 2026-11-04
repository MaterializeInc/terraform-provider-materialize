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

# A second index on the same object and in the same cluster. Of two such
# indexes, the one created later reads from the other, so it needs the same
# block pointing at the other index. The block also makes Terraform create it
# second. Without it, Terraform creates both at the same time, and either can
# end up reading from the other. Adding the block to both indexes is rejected
# as a cycle. For two indexes that already exist, add it to the one that the
# in-use error names.
resource "materialize_index" "loadgen_index_by_id" {
  name         = "example_index_by_id"
  cluster_name = "cluster"

  obj_name {
    name          = "source"
    schema_name   = "schema"
    database_name = "database"
  }

  col_expr {
    field = "id"
  }

  lifecycle {
    replace_triggered_by = [materialize_index.loadgen_index.id]
  }
}

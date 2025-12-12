data "materialize_source_reference" "source_references" {
  source_id = materialize_source_mysql.test.id
}

output "source_references" {
  value = data.materialize_source_reference.my_source_references.references
}

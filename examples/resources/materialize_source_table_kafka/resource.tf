resource "materialize_source_table_kafka" "kafka_source_table" {
  name          = "kafka_source_table"
  schema_name   = "public"
  database_name = "materialize"

  source {
    name          = materialize_source_kafka.test_source_kafka.name
    schema_name   = materialize_source_kafka.test_source_kafka.schema_name
    database_name = materialize_source_kafka.test_source_kafka.database_name
  }

  upstream_name           = "terraform" # The kafka source topic name
  include_key             = true
  include_key_alias       = "message_key"
  include_headers         = true
  include_headers_alias   = "message_headers"
  include_partition       = true
  include_partition_alias = "message_partition"
  include_offset          = true
  include_offset_alias    = "message_offset"
  include_timestamp       = true
  include_timestamp_alias = "message_timestamp"


  key_format {
    text = true
  }
  value_format {
    json = true
  }

  envelope {
    upsert = true
    upsert_options {
      value_decoding_errors {
        inline {
          enabled = true
          alias   = "decoding_error"
        }
      }
    }
  }

  ownership_role = "mz_system"
  comment        = "This is a test Kafka source table"
}

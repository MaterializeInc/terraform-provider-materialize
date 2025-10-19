resource "materialize_sink_kafka" "sink_kafka" {
  name             = "sink_kafka"
  schema_name      = materialize_schema.schema.name
  database_name    = materialize_database.database.name
  comment          = "sink comment"
  cluster_name     = materialize_cluster.cluster_sink.name
  topic            = "topic1"
  key              = ["id"]
  key_not_enforced = true
  from {
    name          = "accounts"
    database_name = materialize_source_load_generator.load_generator_cluster.database_name
    schema_name   = materialize_source_load_generator.load_generator_cluster.schema_name
  }
  kafka_connection {
    name          = materialize_connection_kafka.kafka_connection.name
    database_name = materialize_connection_kafka.kafka_connection.database_name
    schema_name   = materialize_connection_kafka.kafka_connection.schema_name
  }
  format {
    avro {
      schema_registry_connection {
        name          = materialize_connection_confluent_schema_registry.schema_registry.name
        database_name = materialize_connection_confluent_schema_registry.schema_registry.database_name
        schema_name   = materialize_connection_confluent_schema_registry.schema_registry.schema_name
      }
      avro_doc_type {
        object {
          name          = materialize_source_load_generator.load_generator.name
          database_name = materialize_source_load_generator.load_generator.database_name
          schema_name   = materialize_source_load_generator.load_generator.schema_name
        }
        doc = "top level comment"
      }
      avro_doc_column {
        object {
          name          = "accounts"
          database_name = materialize_source_load_generator.load_generator_cluster.database_name
          schema_name   = materialize_source_load_generator.load_generator_cluster.schema_name
        }
        column = "id"
        doc    = "comment key"
        key    = true
      }
      avro_doc_column {
        object {
          name          = "accounts"
          database_name = materialize_source_load_generator.load_generator_cluster.database_name
          schema_name   = materialize_source_load_generator.load_generator_cluster.schema_name
        }
        column = "id"
        doc    = "comment value"
        value  = true
      }
      key_compatibility_level   = "BACKWARD"
      value_compatibility_level = "FORWARD"
    }
  }
  envelope {
    debezium = true
  }
}

resource "materialize_sink_kafka" "sink_kafka_cluster" {
  name             = "sink_kafka_cluster"
  schema_name      = materialize_schema.schema.name
  database_name    = materialize_database.database.name
  cluster_name     = materialize_cluster.cluster_sink.name
  topic            = "topic1"
  key              = ["id"]
  key_not_enforced = true
  snapshot         = true
  from {
    name          = "accounts"
    database_name = materialize_source_load_generator.load_generator_cluster.database_name
    schema_name   = materialize_source_load_generator.load_generator_cluster.schema_name
  }
  kafka_connection {
    name          = materialize_connection_kafka.kafka_connection.name
    database_name = materialize_connection_kafka.kafka_connection.database_name
    schema_name   = materialize_connection_kafka.kafka_connection.schema_name
  }
  format {
    json = true
  }
  envelope {
    upsert = true
  }
}


resource "materialize_sink_kafka" "sink_kafka_headers" {
  name                     = "sink_kafka_headers"
  schema_name              = materialize_schema.schema.name
  database_name            = materialize_database.database.name
  cluster_name             = materialize_cluster.cluster_sink.name
  topic                    = "topic1"
  topic_replication_factor = 1
  topic_partition_count    = 6
  topic_config = {
    "cleanup.policy" = "compact"
    "retention.ms"   = "86400000"
  }
  key              = ["key_column"]
  key_not_enforced = true
  snapshot         = true
  headers          = "kafka_header"
  from {
    name          = materialize_table.simple_table_sink.name
    database_name = materialize_table.simple_table_sink.database_name
    schema_name   = materialize_table.simple_table_sink.schema_name
  }
  kafka_connection {
    name          = materialize_connection_kafka.kafka_connection.name
    database_name = materialize_connection_kafka.kafka_connection.database_name
    schema_name   = materialize_connection_kafka.kafka_connection.schema_name
  }
  format {
    json = true
  }
  envelope {
    upsert = true
  }
}

output "qualified_sink_kafka" {
  value = materialize_sink_kafka.sink_kafka.qualified_sql_name
}

data "materialize_sink" "all" {}

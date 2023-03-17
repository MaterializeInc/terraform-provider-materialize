# Create a Kafka Connection
resource "materialize_connection_kafka" "example_kafka_connection" {
  name = "example_kafka_connection"
  kafka_broker {
    broker = "b-1.hostname-1:9096"
  }
  sasl_username = "example"
  sasl_password {
    name          = "kafka_password"
    database_name = "materialize"
    schema_name   = "public"
  }
  sasl_mechanisms = "SCRAM-SHA-256"
  progress_topic  = "example"
}

# CREATE CONNECTION database.schema.kafka_conn TO KAFKA (
#     BROKER 'example:9092'
#     PROGRESS TOPIC 'topic',
#     SASL MECHANISMS 'PLAIN',
#     SASL USERNAME 'user',
#     SASL PASSWORD SECRET "materialize"."public"."kafka_password"
# );

resource "materialize_connection_kafka" "example_kafka_connection_multiple_brokers" {
  name = "example_kafka_connection_multiple_brokers"
  kafka_broker {
    broker            = "b-1.hostname-1:9096"
    target_group_port = "9001"
    availability_zone = "use1-az1"
    privatelink_connection {
      name          = "example_aws_privatelink_conn"
      database_name = "materialize"
      schema_name   = "public"
    }
  }
  kafka_broker {
    broker            = "b-2.hostname-2:9096"
    target_group_port = "9002"
    availability_zone = "use1-az2"
    privatelink_connection {
      name          = "example_aws_privatelink_conn"
      database_name = "materialize"
      schema_name   = "public"
    }
  }
}

# CREATE CONNECTION materialize.public.example_kafka_connection_multiple_brokers TO KAFKA (
#     BROKERS (
#        'b-1.hostname-1:9096' USING AWS PRIVATELINK "materialize"."public"."example_aws_privatelink_conn" (PORT 9001, AVAILABILITY ZONE 'use1-az1'),
#        'b-2.hostname-2:9096' USING AWS PRIVATELINK "materialize"."public"."example_aws_privatelink_conn" (PORT 9002, AVAILABILITY ZONE 'use1-az2')
#     )
# );

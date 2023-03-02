# Create a Kafka Connection
resource "materialize_connection_kafka" "example_kafka_connection" {
  name            = "example_kafka_connection"
  kafka_broker {
    broker = "b-1.hostname-1:9096"
  }
  kafka_sasl_username   = "example"
  kafka_sasl_password   = "kafka_password"
  kafka_sasl_mechanisms = "SCRAM-SHA-256"
  kafka_progress_topic  = "example"
}

# CREATE CONNECTION database.schema.kafka_conn TO KAFKA (
#     BROKER 'example:9092'
#     PROGRESS TOPIC 'topic',
#     SASL MECHANISMS 'PLAIN',
#     SASL USERNAME 'user',
#     SASL PASSWORD SECRET password
# );

resource "materialize_connection_kafka" "example_kafka_connection_multiple_brokers" {
  name            = "example_kafka_connection_multiple_brokers"
  kafka_broker {
    broker = "b-1.hostname-1:9096"
    target_group_port = "9001"
    availability_zone = "use1-az1"
    privatelink_connection = "example_aws_privatelink_conn"
  }
  kafka_broker {
    broker = "b-2.hostname-2:9096"
    target_group_port = "9002"
    availability_zone = "use1-az2"
    privatelink_connection = "example_aws_privatelink_conn"
  }
}

# CREATE CONNECTION materialize.public.example_kafka_connection_multiple_brokers TO KAFKA (
#     BROKERS (
#        'b-1.hostname-1:9096' USING AWS PRIVATELINK example_aws_privatelink_conn (PORT 9001, AVAILABILITY ZONE 'use1-az1'),
#        'b-2.hostname-2:9096' USING AWS PRIVATELINK example_aws_privatelink_conn (PORT 9002, AVAILABILITY ZONE 'use1-az2')
#     )
# );

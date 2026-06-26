# Create a Kafka Connection
resource "materialize_connection_kafka" "example_kafka_connection" {
  name = "example_kafka_connection"
  kafka_broker {
    broker = "b-1.hostname-1:9096"
  }
  sasl_username {
    text = "user"
  }
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

# Kafka Connection using AWS PrivateLink with wildcard broker matching rules.
# Useful for clusters with dynamic broker discovery (e.g. Confluent Cloud):
# a static broker bootstraps the connection, and MATCHING rules route the
# brokers Kafka advertises through the correct per-AZ PrivateLink endpoint.
# Requires the `enable_kafka_broker_matching_rules` feature in your region.
resource "materialize_connection_kafka" "example_kafka_connection_matching_rules" {
  name = "example_kafka_connection_matching_rules"
  kafka_broker {
    broker = "lkc-825730.domain.confluent.cloud:9092"
    privatelink_connection {
      name          = "example_aws_privatelink_conn"
      database_name = "materialize"
      schema_name   = "public"
    }
  }
  broker_matching_rule {
    pattern           = "*.use1-az1.*"
    availability_zone = "use1-az1"
    privatelink_connection {
      name          = "example_aws_privatelink_conn"
      database_name = "materialize"
      schema_name   = "public"
    }
  }
  broker_matching_rule {
    pattern           = "*.use1-az4.*"
    availability_zone = "use1-az4"
    privatelink_connection {
      name          = "example_aws_privatelink_conn"
      database_name = "materialize"
      schema_name   = "public"
    }
  }
}

# CREATE CONNECTION materialize.public.example_kafka_connection_matching_rules TO KAFKA (
#     BROKERS (
#        'lkc-825730.domain.confluent.cloud:9092' USING AWS PRIVATELINK "materialize"."public"."example_aws_privatelink_conn",
#        MATCHING '*.use1-az1.*' USING AWS PRIVATELINK "materialize"."public"."example_aws_privatelink_conn" (AVAILABILITY ZONE 'use1-az1'),
#        MATCHING '*.use1-az4.*' USING AWS PRIVATELINK "materialize"."public"."example_aws_privatelink_conn" (AVAILABILITY ZONE 'use1-az4')
#     )
# );

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSourceKafkaMigration_basic(t *testing.T) {
	addTestTopic()
	sourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"materialize": {
						VersionConstraint: "0.4.1",
						Source:            "MaterializeInc/materialize",
					},
				},
				Config: testAccSourceKafkaMigrationV0Resource(sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceKafkaExists("materialize_source_kafka.test"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "start_offset.#", "1"),
				),
			},
			{
				ProviderFactories: testAccProviderFactories,
				Config:            testAccSourceKafkaMigrationV1Resource(sourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceKafkaExists("materialize_source_kafka.test"),
					resource.TestCheckResourceAttr("materialize_source_kafka.test", "start_offsets.#", "1"),
				),
			},
			{
				ProviderFactories: testAccProviderFactories,
				ResourceName:      "materialize_source_kafka.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func testAccSourceKafkaMigrationV0Resource(sourceName string) string {
	return fmt.Sprintf(`
	resource "materialize_connection_kafka" "test" {
		name = "%[1]s"
		kafka_broker {
			broker = "redpanda:9092"
		}
		security_protocol = "PLAINTEXT"
	}

	resource "materialize_source_kafka" "test" {
		name = "%[1]s"
		kafka_connection {
			name = materialize_connection_kafka.test.name
		}

		size  = "3xsmall"
		topic = "terraform"
		key_format {
			text = true
		}
		value_format {
			text = true
		}
		envelope {
			none = true
		}
		start_offset = [0]
	}
`, sourceName)
}

func testAccSourceKafkaMigrationV1Resource(sourceName string) string {
	return fmt.Sprintf(`
	resource "materialize_connection_kafka" "test" {
		name = "%[1]s"
		kafka_broker {
			broker = "redpanda:9092"
		}
		security_protocol = "PLAINTEXT"
	}

	resource "materialize_source_kafka" "test" {
		name = "%[1]s"
		kafka_connection {
			name = materialize_connection_kafka.test.name
		}

		size  = "3xsmall"
		topic = "terraform"
		key_format {
			text = true
		}
		value_format {
			text = true
		}
		envelope {
			none = true
		}
		start_offset = [0]
	}
`, sourceName)
}

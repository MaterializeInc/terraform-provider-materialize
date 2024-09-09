# Cloud Mock API

This mock service provides a configurable API for simulating the Materialize cloud API endpoints.

## Prerequisites

- Go 1.22 or later
- Docker

## Configuration

The mock API can be configured using environment variables. Here are the available configuration options:

- `CLOUD_HOSTNAME`: Hostname for the cloud API (default: "localhost")
- `CLOUD_PORT`: Port for the cloud API (default: "3001")
- `US_EAST_1_HOSTNAME`: Hostname for the US East 1 region (default: "materialized")
- `US_EAST_1_SQL_PORT`: SQL port for the US East 1 region (default: "6877")
- `US_EAST_1_HTTP_PORT`: HTTP port for the US East 1 region (default: "6875")
- `US_WEST_2_HOSTNAME`: Hostname for the US West 2 region (default: "materialized2")
- `US_WEST_2_SQL_PORT`: SQL port for the US West 2 region (default: "7877")
- `US_WEST_2_HTTP_PORT`: HTTP port for the US West 2 region (default: "7875")

## Running the Mock API

1. Build the Docker image:
   ```
   docker build -t cloud-mock-api .
   ```

2. Run the container:
   ```
   docker run -p 3001:3001 \
     -e CLOUD_HOSTNAME=cloud \
     -e CLOUD_PORT=3001 \
     -e US_EAST_1_HOSTNAME=materialized \
     -e US_EAST_1_SQL_PORT=6877 \
     -e US_EAST_1_HTTP_PORT=6875 \
     -e US_WEST_2_HOSTNAME=materialized2 \
     -e US_WEST_2_SQL_PORT=7877 \
     -e US_WEST_2_HTTP_PORT=7875 \
     cloud-mock-api
   ```

### Docker Compose Setup

You can also use Docker Compose for easier setup and management. Create a `docker-compose.yml` file with the following content:

```yaml
version: '3'
services:
  cloud:
    build: .
    ports:
      - "3001:3001"
    environment:
      - CLOUD_HOSTNAME=cloud
      - CLOUD_PORT=3001
      - US_EAST_1_HOSTNAME=materialized
      - US_EAST_1_SQL_PORT=6877
      - US_EAST_1_HTTP_PORT=6875
      - US_WEST_2_HOSTNAME=materialized2
      - US_WEST_2_SQL_PORT=7877
      - US_WEST_2_HTTP_PORT=7875
```

Then run:

```
docker compose up -d --build
```

## API Endpoints

- `/api/cloud-regions`: Get information about all cloud regions
- `/{region-name}/api/region`: Get, update, or delete information about a specific region (e.g., `/us-east-1/api/region`).

## Using with Terraform

To use this mock API with Terraform, configure your provider like this:

```hcl
provider "materialize" {
  # Define the cloud endpoint and port
  cloud_endpoint = "http://${var.cloud_hostname}:${var.cloud_port}"

  endpoint       = "http://0.0.0.0:3000"
  password       = "your_app_password_here"
  sslmode        = "disable"
  database       = "materialize"
}
```

Replace `var.cloud_hostname` and `var.cloud_port` with the appropriate values for your setup.

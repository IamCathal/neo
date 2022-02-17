# Services

All neo microservices

## Configuration

All services must have the following environment variables

| Variable     | Description |
| ----------- | ----------- |
| `API_PORT`      | The port to host the application on       |
| `LOG_PATH`   | Where to write logs to        |
| `NODE_NAME`   | Unique name for this instance      |
| `NODE_DC`   | Datacenter for this instance        |
| `SYSTEM_STATS_BUCKET`   | Bucket name for system stats metrics        |
| `SYSTEM_STATS_BUCKET_TOKEN`   | Bucket token for system stats metrics        |
| `ENDPOINT_LATENCIES_BUCKET`   | Bucket name for endpoint latencies        |
| `ENDPOINT_LATENCIES_BUCKET_TOKEN`   | Bucket token for endpoint latencies metrics        |
| `ORG`   | Org name for grafana        |
| `INFLUXDB_URL`   | Full URL for connecting to influxDB        |
| `AUTH_KEY`     | Authentication key used by services for authorized only endpoints |

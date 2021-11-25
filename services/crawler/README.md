# Crawler

![test status badge](https://github.com/IamCathal/neo/actions/workflows/buildCrawler.yml/badge.svg)

The crawler service asynchronously consumes jobs from rabbitMQ for crawling and outputs the user's crawled profile information to long term storage.

## Configuration

Crawler instances expect the following variables to be set in .env

#### Crawler specific

| Variable     | Description |
| ----------- | ----------- |
| `WORKER_AMOUNT` | Number of workers to run per node    |
| `RABBITMQ_QUEUE_NAME` | Name of the rabbitMQ queue   |
| `RABBITMQ_USER` | RabbitMQ username    |
| `RABBITMQ_URL` | URL (port included) of RabbitmQ instance    |
| `DATASTORE_URL` | URL (port included) of the datastore instance    |
| `STEAM_API_KEYS` | Comma seperated list of Steam web API keys    |
| `KEY_USAGE_TIMER` | Minimum time elapsed in milliseconds between subsequent uses of a given API key    |

#### Default

| Variable     | Description |
| ----------- | ----------- |
| `API_PORT`      | The port to host the application on       |
| `LOG_PATH`   | Where to write logs to        |
| `NODE_NAME`   | Unique name for this instance      |
| `NODE_DC`   | Datacenter for this instance        |
| `SYSTEM_STATS_BUCKET`   | Bucket name for system stats metrics        |
| `SYSTEM_STATS_BUCKET_TOKEN`   | Bucket token for system stats metrics        |
| `ORG`   | Org name for grafana        |
| `INFLUXDB_URL`   | Full URL for connecting to influxDB        |

## Running 

`docker-compose up` to start with docker-compose (preferred)

`docker build -f Dockerfile -t iamcathal/crawler:0.0.1 .` and `docker run -it --rm -p PORT:PORT iamcathal/crawler:0.0.1` to start as a standalone container
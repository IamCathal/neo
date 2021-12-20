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

## Running 

`docker-compose up` to start with docker-compose (preferred)

`docker build -f Dockerfile -t iamcathal/crawler:0.0.1 .` and `docker run -it --rm -p PORT:PORT iamcathal/crawler:0.0.1` to start as a standalone container
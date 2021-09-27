# Crawler

The crawler service asynchronously consumes jobs from rabbitMQ for crawling and outputs the user's crawled profile information to long term storage.

## Configuration

Crawler instances expect the following variables to be set in .env

#### Crawler specific

Crawler does not currently require any specific environtment variables.

#### Default

| Variable     | Description |
| ----------- | ----------- |
| `API_PORT`      | The port to host the application on       |
| `LOG_PATH`   | Where to write logs to        |
| `NODE_NAME`   | Unique name for this instance      |
| `NODE_DC`   | Datacenter for this instance        |

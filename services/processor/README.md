# Processor

The processor service aggregrates all of the crawled data for a given user and generates a html graph output that the user is presented with

## Configuration

Processor instances expect the following variables to be set in .env

#### Processor specific

Processor does not currently require any specific environtment variables.

#### Default

| Variable     | Description |
| ----------- | ----------- |
| `API_PORT`      | The port to host the application on       |
| `LOG_PATH`   | Where to write logs to        |
| `NODE_NAME`   | Unique name for this instance      |
| `NODE_DC`   | Datacenter for this instance        |

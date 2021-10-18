# Frontend

![frontend test status badge](https://github.com/IamCathal/neo/actions/workflows/buildFrontEnd.yml/badge.svg)   ![frontend deploy status](https://github.com/IamCathal/neo/actions/workflows/deployFrontend.yml/badge.svg) 

Frontend of neo

## Configuration

Frontend expects the following variables to be set in .env

#### Frontend specific

| Variable     | Description |
| ----------- | ----------- |
| `STATIC_CONTENT_DIR_NAME`      | Name of the static content directory i.e `./static`     |


#### Default

| Variable     | Description |
| ----------- | ----------- |
| `API_PORT`      | The port to host the application on       |
| `LOG_PATH`   | Where to write logs to        |
| `NODE_NAME`   | Unique name for this instance      |
| `NODE_DC`   | Datacenter for this instance        |

## Running 

`docker-compose up` to start with docker-compose (preferred)

`docker build -f Dockerfile -t iamcathal/crawler:0.0.1 .` and `docker run -it --rm -p PORT:PORT iamcathal/crawler:0.0.1` to start as a standalone container
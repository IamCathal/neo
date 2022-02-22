# Frontend

![frontend test status badge](https://github.com/IamCathal/neo/actions/workflows/buildFrontEnd.yml/badge.svg)   ![frontend deploy status](https://github.com/IamCathal/neo/actions/workflows/deployFrontend.yml/badge.svg) 

Frontend of neo

## Configuration

Frontend expects the following variables to be set in .env

#### Frontend specific

| Variable     | Description |
| ----------- | ----------- |
| `STATIC_CONTENT_DIR_NAME`      | Name of the static content directory i.e `./static`     |
| `CRAWLER_INSTANCE`      | URL (port included) for any crawler instance     |
| `DATASTORE_INSTANCE` | URL (port included) of the datastore instance    |

## Running 

`docker-compose up` to start with docker-compose (preferred)

`docker build -f Dockerfile -t iamcathal/crawler:0.0.1 .` and `docker run -it --rm -p PORT:PORT iamcathal/crawler:0.0.1` to start as a standalone container
# Datastore

![datastore test status badge](https://github.com/IamCathal/neo/actions/workflows/buildDatastore.yml/badge.svg)   ![datastore deploy status]() 

Datastore is the thin client that sits infront of the database

## Configuration

Datastore expects the following variables to be set in .env

#### Datastore specific

| Variable     | Description |
| ----------- | ----------- |
| `MONGODB_USER`      |  MongoDB account username  |
| `MONGODB_PASSWORD`      |  MongoDB account password  |
| `MONGO_INSTANCE_IP`      |  IP for the MongoDB instance |
| `DB_NAME`      |  Database name for the stored data |
| `USER_COLLECTION`      |  Collection name for the user data |
| `CRAWLING_STATS_COLLECTION`      |  Collection name for the crawling stats |



#### Default

| Variable     | Description |
| ----------- | ----------- |
| `API_PORT`      | The port to host the application on       |
| `LOG_PATH`   | Where to write logs to        |
| `NODE_NAME`   | Unique name for this instance      |
| `SYSTEM_STATS_BUCKET`   | Bucket name for system stats metrics        |
| `SYSTEM_STATS_BUCKET_TOKEN`   | Bucket token for system stats metrics        |
| `ORG`   | Org name for grafana        |
| `INFLUXDB_URL`   | Full URL for connecting to influxDB        |


## Running 

`docker-compose up` to start with docker-compose (preferred)

`docker build -f Dockerfile -t iamcathal/datastore:0.0.1 .` and `docker run -it --rm -p PORT:PORT iamcathal/datastore:0.0.1` to start as a standalone container
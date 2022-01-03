# Postgresql

Simple single table postgres instance used for storing and serving users processed graph data

## Configuration

The following env vars are expected by postgres:

| Variable     | Description |
| ----------- | ----------- |
| `POSTGRES_USER`      |  Username  |
| `POSTGRES_PASSWORD`      |  Password  |
| `POSTGRES_DB`      |  Default DB name |

On localhost the instance can be connected to using psql using `psql -d dbName -U username -W`
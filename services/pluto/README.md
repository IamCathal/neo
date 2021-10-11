# Pluto

![test status badge](https://github.com/IamCathal/neo/actions/workflows/buildPluto.yml/badge.svg)

Pluto receives webhooks for events happening on github actions and sends prompts to discord to deploy different services once they've successfully built.

## Configuration

Pluto expects the following variables to be set in .env

#### Pluto specific

| Variable     | Description |
| ----------- | ----------- |
| `WEBHOOK_SECRET`   | Secret used to verify authenticity of Github webhooks        |
| `BOT_TOKEN`   | Discord bot token       |
| `DISCORD_CHANNEL_ID`   | Channel ID of the deploy channel on discord      |


#### Default

| Variable     | Description |
| ----------- | ----------- |
| `API_PORT`      | The port to host the application on       |
| `LOG_PATH`   | Where to write logs to        |
| `NODE_NAME`   | Unique name for this instance      |
| `NODE_DC`   | Datacenter for this instance        |

## Running 

`docker-compose up` to start with docker-compose (preferred)

`docker build -f Dockerfile -t iamcathal/pluto:0.0.1 .` and `docker run -it --rm -p PORT:PORT iamcathal/pluto:0.0.1` to start as a standalone container
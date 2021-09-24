# Pluto

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

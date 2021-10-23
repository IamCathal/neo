# Grafana and InfluxDB

This docker-compose runs the grafana and influxDB.

## Configuration
 
* Setup InfluxDB

Go to xxx.xxx:8086 and create a new user with initial bucket etc. Load Daaa > Tokens and create a token for metrics pushing from nodes with only write access. Create a token with read and write access for all buckets for grafana to use

* Setup Grafana

Configuration > Data Sources and add influxDB. Choose flux as query language and use correct login details

# Setup Grafana
# Health check endpoints

### Grafana

* GET `http://localhost:3000/api/health`

### InfluxDB

* GET `http://localhost:8086/health`

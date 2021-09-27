# RabbitMQ

This docker-compose runs rabbitMQ and prometheus for metric aggregration

# Configuration

* Guest user must be deleted or have password changed. [Manage users](http://localhost:15672/#/users).
*-* When sending API requests to authentication endpoints input username and password in the form `http://user:pass@localhost:15672`/ See [here](https://serverfault.com/a/371918) for explanation

# Health check endpoints

### RabbitMQ

* GET `http://localhost:15672/`

### InfluxDB

* GET `http://localhost:9090/-/healthy`
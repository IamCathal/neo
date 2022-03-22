<p align="center">
 <img width="20%" src="https://i.imgur.com/3DH7xYd.png">
</p>


# neo

NUIG CS&IT Final Year Project 2022


| Service     | Tests | Deploy |  
| ----------- | ----------- | ----------- |
| Frontend      | ![frontend test status badge](https://github.com/IamCathal/neo/actions/workflows/buildFrontEnd.yml/badge.svg) |  ![frontend deploy status](https://github.com/IamCathal/neo/actions/workflows/deployFrontend.yml/badge.svg) |  
| Datastore | ![datastore test status badge](https://github.com/IamCathal/neo/actions/workflows/buildDatastore.yml/badge.svg) | ![datastore deploy status](https://github.com/IamCathal/neo/actions/workflows/deployDataStore.yml/badge.svg)  |
| Crawler      | ![test test status badge](https://github.com/IamCathal/neo/actions/workflows/buildCrawler.yml/badge.svg) | ![crawler deploy status](https://github.com/IamCathal/neo/actions/workflows/deployCrawler.yml/badge.svg)  |  

## Where to get started?

Browse the main microservices and pieces of infrastructure

### Microservices

* [Frontend](services/frontend)
* [Crawler](services/crawler)
* [DataStore](services/data)

### Infrastructure

* [Logging (ELK Stack)](infrastructure/elk)
* [Metrics (Grafana, InfluxDB)](infrastructure/grafana)
* [MongoDB](infrastructure/mongoDB)
* [PostgreSQL](infrastructure/postgresql)
* [Queue/Messaging System (RabbitMQ)](infrastructure/rabbitMQ)



## Architecture

![Architecture diagram](services/frontend/static/images/NeoArchitectureFinal.png)

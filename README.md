<p align="center">
 <a href="http://neofyp.com">
 <img width="20%" src="https://i.imgur.com/3DH7xYd.png">
 </a>
</p>


# neo


Generate an indepth analysis of your friend network on Steam to see exactly what kind of friends you have or find the shortest distance between two steam users. This project is my NUIG CS&IT Final Year Project 2022.

## Deployment

Neo is a distributed system deployed across 9 servers hosted mostly by OVH. V̶i̶s̶i̶t̶ ̶t̶h̶e̶ ̶d̶e̶p̶l̶o̶y̶e̶d̶ ̶p̶r̶o̶j̶e̶c̶t̶ ̶a̶t̶ A static snapshot of the site can be viewed at [cathaloc.dev/fyp](https://cathaloc.dev/fyp)

| Service     | Tests | Deploy |  
| ----------- | ----------- | ----------- |
| Frontend      | ![frontend test status badge](https://github.com/IamCathal/neo/actions/workflows/buildFrontEnd.yml/badge.svg) |  ![frontend deploy status](https://github.com/IamCathal/neo/actions/workflows/deployFrontend.yml/badge.svg) |  
| Datastore | ![datastore test status badge](https://github.com/IamCathal/neo/actions/workflows/buildDatastore.yml/badge.svg) | ![datastore deploy status](https://github.com/IamCathal/neo/actions/workflows/deployDataStore.yml/badge.svg)  |
| Crawler      | ![test test status badge](https://github.com/IamCathal/neo/actions/workflows/buildCrawler.yml/badge.svg) | ![crawler deploy status](https://github.com/IamCathal/neo/actions/workflows/deployCrawler.yml/badge.svg)  |  


## Architecture

![Architecture diagram](services/frontend/static/images/NeoArchitectureFinal.png)

# Gallery

| Landing Page    | Section of Shortest Distance Page | 
| ----------- | ----------- | 
|      ![](https://i.imgur.com/YFnBviP.png)       |    ![](https://i.imgur.com/xTHPqXD.png)  |

| Section of Graph Page    | Interactive Graph Page | 
| ----------- | ----------- | 
|      ![](https://i.imgur.com/HR9UEqk.jpeg)       |    ![](https://i.imgur.com/MJ4Lkvi.png)  |


# Documentation

My submitted FYP report detailing everything from planning, development and operations can be viewed [here](https://cathaloc.dev/static/pdfs/fyp.pdf)

## Where to get started?

If you're looking to browse the codebase here are the main points of interest:

### Microservices

* [Frontend](services/frontend)
* [Crawler](services/crawler)
* [DataStore](services/datastore)

### Infrastructure

* [Logging (ELK Stack)](infrastructure/elk)
* [Metrics (Grafana, InfluxDB)](infrastructure/grafana)
* [Database (MongoDB)](infrastructure/mongoDB)
* [Database (PostgreSQL)](infrastructure/postgresql)
* [Queue/Messaging System (RabbitMQ)](infrastructure/rabbitMQ)

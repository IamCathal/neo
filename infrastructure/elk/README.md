# Elasticsearch, Logstash & Kibana

This docker-compose runs elasticsearch, logstash and kibana for log processing and browsing.

# Configuration

Default usernames and passwords in setup yml files should be changed

# Health check endpoints

### Elasticsearch 

* GET `http://localhost:9200/_cat/health`

### Logstash

* GET `http://localhost:9600/?pretty`

### Kibana

* GET `http://localhost:5601/login`
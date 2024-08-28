## Klaunch utility

This is a CLI tool to perform MongoDB Kafka connect reproductions.

## Pre requisites

- mlaunch installed.
- 3 node replicased running on localy.

##  Commands

- start [connector version]: Creates a Docker compose with all the necesary infrastrusture components.
By default will connect to release repository and download the latest version of MongoDB Kafka Connect.

- stop: Deletes the Docker compose completely.

- create: Creates a connector Task based on an input config file path.(json format) 

- delete: Deletes all existing Tasks and topics. Infrastructure remains.

- show [components - messages]
    - Components: will list running Tasks and exisiting Topics.
    - Messages: will list existing Topics and will create a consumer process to display messages on the console.

- logs: Will dump a the Kafka connect log file into $repository/logs path with the following format: `$timestamps_kadka_connect.log`



-  [MongoDB Kafka Connector](https://docs.mongodb.com/kafka-connector/current/) & Kafka repro environment easy to build & destroy.
- Visual monitoring system - Console Management 
  The Monitoring expresses these logs visually, to make analyzing the system more straightforward providing the following monitoring:
  - [MongoDB Kafka Connector](https://docs.mongodb.com/kafka-connector/current/) metrics on Kafka Connect Grafana dashboard.
  - Kafka broker & Zookeeper Grafana dashboards.
  - Monitoring database on Prometheus.

- Communication between Docker and `mlaunch` repro tool.
    - TCP/IP communication between a docker container to localhost MongoDB Replica Set or `mongod` service.

**[MongoDB Kafka Connector](https://docs.mongodb.com/kafka-connector/current/) Monitoring dashboard examples:**

### Components

- [Docker](https://www.docker.com/) is a set of products that use OS-level virtualization to deliver software in packages called containers.
- [Apache Kafka](https://kafka.apache.org/) is a framework implementation of a software bus using stream-processing
- [Cluster Manager for Apache Kafka](https://github.com/yahoo/CMAK) is a tool for managing Apache Kafka clusters.
- [Zookeeper](https://zookeeper.apache.org/) is a centralized service for maintaining configuration and naming.

---------------------------------


### Disclaimer

> This project uses code from other sources.
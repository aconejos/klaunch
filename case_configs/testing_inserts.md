-- Deploy connector 
curl -s -X POST -H 'Content-Type: application/json'  \
--data @./case_configs/default_topic.json http://localhost:8083/connectors


-- Delete connector
curl -X DELETE  http://localhost:8083/connectors/mdb-kafka-connector-default

-- Delete topic

docker exec kafka-connect kafka-topics \
  --delete \
  --bootstrap-server kafka2:19092,kafka3:19093,kafka1:19091 \
  --topic mdb_kafka_test.connector_test



-- insert test data

use mdb_kafka_test

db.connector_test.insertOne({name: "Example Document", description: "This is an example document1"})
db.connector_test.insertOne({name: "Example Document", description: "This is an example document2"})
db.connector_test.insertOne({name: "Example Document", description: "This is an example document3"})
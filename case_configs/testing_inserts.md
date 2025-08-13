-- Deploy connector 
curl -s -X POST -H 'Content-Type: application/json' --data @./case_configs/default_topic.json http://localhost:8083/connectors

curl -s -X GET http://localhost:8083/connectors \
-H 'Content-Type: application/json' 

-- Delete connector
curl -X DELETE  http://localhost:8083/connectors/mdb-kafka-connector-default

-- Delete topic

docker exec kafka-connect kafka-topics \
  --delete \
  --bootstrap-server kafka2:19092,kafka3:19093,kafka1:19091 \
  --topic mdb_kafka_test.connector_test

-- track changes

docker exec kafka-connect kafka-console-consumer --topic disable.db_name.coll_name --from-beginning --max-messages 10 --bootstrap-server=kafka2:19092,kafka3:19093,kafka1:19091

-- insert test data

use mdb_kafka_test

db.connector_test.insertOne({name: "Example Document", description: "This is an example document1"})
db.connector_test.insertOne({name: "Example Document", description: "This is an example document2"})
db.connector_test.insertOne({name: "Example Document", description: "This is an example document3"})

-- simple source connector

use Tutorial1
db.orders.insertOne( { 'order_id' : 1, 'item' : 'coffee' } )


db.collectionName.insertOne({  
  id: UUID(),                          // Generate a new UUID  
  type: "your_type_value",             // Replace with your desired type  
  value: "your_value",                 // Replace with your desired value  
  location: "Barcelona",               // Fixed to 'Barcelona'  
  createdAt: new Date()                // Automatically set to the current timestamp  
}); 


`curl -i -X PUT -H "Content-Type: application/json" \http://localhost:8083/connectors/connector_name1/config \-d '{
             "connector.class":"com.mongodb.kafka.connect.MongoSourceConnector",
             "tasks.max":10,
             "connection.uri": "mongodb://host.docker.internal:27017,host.docker.internal:27018,host.docker.internal:27019/replicaSet=replset",
             "database":"db_name",
             "collection":"coll_name",
             "startup.mode":"copy_existing",
             "pipeline":"[{\"$match\":{\"fullDocument.knp\":true}},{\"$project\":{\"fullDocument.eventID\":1, \"ns\":1}}]",
             "topic.prefix":"testing",
             "topic.namespace.map":"{\"namespace\":\"topic_name\"}",
             "poll.max.batch.size":1000,
             "poll.await.time.ms":5000,
             "publish.full.document.only":false,
             "publish.full.document.only.tombstone.on.delete":false,
             "change.stream.full.document":"updateLookup",
             "key.converter" : "org.apache.kafka.connect.storage.StringConverter",
             "value.converter" : "org.apache.kafka.connect.storage.StringConverter",
             "output.format.key":"json",
             "output.format.value":"json",
             "output.json.formatter":"com.mongodb.kafka.connect.source.json.formatter.SimplifiedJson",
             "mongo.errors.log.enable":true

      }'`

-- Test connections

"connection.uri": "mongodb+srv://aconejos:aconejos@cluster0.dbvaign.mongodb.net/",

mongosh "mongodb+srv://cluster0.dbvaign.mongodb.net/" --apiVersion 1 --username aconejos
const { MongoClient } = require('mongodb');
const { Kafka } = require('kafkajs');

// Configuration
const MONGO_URI = 'mongodb://localhost:27017,localhost:27018,localhost:27019/?replicaSet=replset';
const KAFKA_BROKERS = ['localhost:9091', 'localhost:9092', 'localhost:9093'];
const KAFKA_TOPIC = 'source_db_test.source_collection_test';

// MongoDB databases to monitor
const SOURCE_DB = 'source_db_test';
const SINK_DB_MASKED = 'sink_db_test';
const SINK_DB_RENAMED = 'sink_db_test';

class TransformMonitor {
    constructor() {
        // MongoDB connections
        this.sourceClient = new MongoClient(MONGO_URI);
        this.sinkClient = new MongoClient(MONGO_URI);
        
        // Kafka setup
        this.kafka = new Kafka({
            clientId: 'transform-monitor',
            brokers: KAFKA_BROKERS,
        });
        this.consumer = this.kafka.consumer({ groupId: 'monitor-group-' + Date.now() });
        
        this.isRunning = false;
        this.messageCount = 0;
        this.changeStreams = []; // Track change streams for proper cleanup
    }

    async connect() {
        try {
            // Connect to MongoDB
            await this.sourceClient.connect();
            await this.sinkClient.connect();
            console.log('✅ Connected to MongoDB (source and sink)');

            // Connect to Kafka
            await this.consumer.connect();
            await this.consumer.subscribe({ topic: KAFKA_TOPIC, fromBeginning: false });
            console.log('✅ Connected to Kafka and subscribed to topic:', KAFKA_TOPIC);

        } catch (error) {
            console.error('❌ Connection failed:', error);
            throw error;
        }
    }

    async setupChangeStreams() {
        console.log('\n📡 Setting up MongoDB change streams...');
        
        // Monitor source collection
        const sourceCollection = this.sourceClient.db(SOURCE_DB).collection('source_collection_test');
        const sourceChangeStream = sourceCollection.watch([], { fullDocument: 'updateLookup' });
        this.changeStreams.push(sourceChangeStream);
        
        sourceChangeStream.on('change', (change) => {
            if (change.operationType === 'insert') {
                console.log('\n🟢 SOURCE INSERT DETECTED:');
                console.log('Document ID:', change.fullDocument._id);
                console.log('Email fields:', {
                    email: change.fullDocument.email,
                    contact_email: change.fullDocument.contact_email,
                    user_email: change.fullDocument.user_email
                });
                console.log('UserId:', change.fullDocument.userId);
            }
        });

        sourceChangeStream.on('error', (error) => {
            if (!error.message.includes('client is closed')) {
                console.error('❌ Source change stream error:', error);
            }
        });

        // Monitor masked sink collection
        const maskedCollection = this.sinkClient.db(SINK_DB_MASKED).collection('sink_collection_masked');
        const maskedChangeStream = maskedCollection.watch([], { fullDocument: 'updateLookup' });
        this.changeStreams.push(maskedChangeStream);
        
        maskedChangeStream.on('change', (change) => {
            if (change.operationType === 'insert') {
                console.log('\n🔒 MASKED SINK INSERT DETECTED:');
                console.log('Document ID:', change.fullDocument._id);
                console.log('Masked email fields:', {
                    email: change.fullDocument.email,
                    contact_email: change.fullDocument.contact_email,
                    user_email: change.fullDocument.user_email
                });
                console.log('✨ MASKING TRANSFORMATION APPLIED ✨');
            }
        });

        maskedChangeStream.on('error', (error) => {
            if (!error.message.includes('client is closed')) {
                console.error('❌ Masked change stream error:', error);
            }
        });

        // Monitor renamed sink collection
        const renamedCollection = this.sinkClient.db(SINK_DB_RENAMED).collection('sink_collection_renamed');
        const renamedChangeStream = renamedCollection.watch([], { fullDocument: 'updateLookup' });
        this.changeStreams.push(renamedChangeStream);
        
        renamedChangeStream.on('change', (change) => {
            if (change.operationType === 'insert') {
                console.log('\n🔄 RENAMED SINK INSERT DETECTED:');
                console.log('Original _id field:', change.fullDocument._id);
                console.log('Renamed mongodb_id field:', change.fullDocument.mongodb_id);
                console.log('Original userId field:', change.fullDocument.userId);
                console.log('Renamed user_identifier field:', change.fullDocument.user_identifier);
                console.log('✨ FIELD RENAMING TRANSFORMATION APPLIED ✨');
            }
        });

        renamedChangeStream.on('error', (error) => {
            if (!error.message.includes('client is closed')) {
                console.error('❌ Renamed change stream error:', error);
            }
        });

        console.log('📡 Change streams active for source and sink collections');
    }

    async startKafkaConsumer() {
        console.log('\n🎯 Starting Kafka message consumer...');
        
        await this.consumer.run({
            eachMessage: async ({ topic, partition, message }) => {
                this.messageCount++;
                
                try {
                    const messageValue = JSON.parse(message.value.toString());
                    
                    console.log('\n📨 KAFKA MESSAGE RECEIVED:');
                    console.log(`Message #${this.messageCount} | Topic: ${topic} | Partition: ${partition}`);
                    console.log('Operation:', messageValue.operationType || 'Unknown');
                    
                    if (messageValue.fullDocument) {
                        console.log('Document preview:', {
                            _id: messageValue.fullDocument._id,
                            userId: messageValue.fullDocument.userId,
                            email: messageValue.fullDocument.email,
                            name: messageValue.fullDocument.name
                        });
                    }
                    
                    console.log('🔄 Message will be processed by sink connectors with transformations...');
                    
                } catch (error) {
                    console.error('❌ Error parsing Kafka message:', error);
                }
            },
        });
    }

 

    async start() {
        console.log('🚀 Starting Transform Monitor for SMT Demo');
        console.log(`
📋 Monitoring Configuration:
  - Kafka Topic: ${KAFKA_TOPIC}
  - Source DB: ${SOURCE_DB}.source_collection_test  
  - Masked Sink: ${SINK_DB_MASKED}.sink_collection_masked
  - Renamed Sink: ${SINK_DB_RENAMED}.sink_collection_renamed
        `);

        try {
            await this.connect();
            await this.setupChangeStreams();
            
            console.log('\n🎬 Monitoring started! Insert data using scripts/data_producer.js');
            console.log('👀 Watch for transformation effects in real-time...');
            console.log('🔍 Type "compare <document_id>" to compare transformations');
            console.log('🛑 Press Ctrl+C to stop monitoring\n');
            
            // Setup interactive commands
            this.setupInteractiveCommands();
            
            this.isRunning = true;
            await this.startKafkaConsumer();
            
        } catch (error) {
            console.error('❌ Monitor startup failed:', error);
            await this.stop();
        }
    }

    async stop() {
        console.log('\n🛑 Stopping Transform Monitor...');
        this.isRunning = false;
        
        try {
            // Close change streams first
            for (const changeStream of this.changeStreams) {
                try {
                    await changeStream.close();
                } catch (error) {
                    // Ignore errors when closing change streams
                }
            }
            this.changeStreams = [];

            // Disconnect Kafka consumer
            try {
                await this.consumer.disconnect();
            } catch (error) {
                // Ignore Kafka disconnect errors
            }

            // Close readline interface
            if (this.readline) {
                this.readline.close();
            }

            // Close MongoDB connections
            try {
                await this.sourceClient.close();
            } catch (error) {
                // Ignore MongoDB close errors
            }

            try {
                await this.sinkClient.close();
            } catch (error) {
                // Ignore MongoDB close errors
            }

            console.log('✅ All connections closed');
        } catch (error) {
            console.error('❌ Error during shutdown:', error);
        }
    }

    setupInteractiveCommands() {
        const readline = require('readline');
        const rl = readline.createInterface({
            input: process.stdin,
            output: process.stdout
        });

        rl.on('line', async (input) => {
            const [command, ...args] = input.trim().split(' ');
            
            if (command === 'compare' && args.length > 0) {
                const documentId = args.join(' ');
                await this.compareDocuments(documentId);
            } else if (command === 'help') {
                console.log('\n📋 Available commands:');
                console.log('  compare <document_id> - Compare original vs transformed documents');
                console.log('  help - Show this help message');
                console.log('  Ctrl+C - Exit monitor\n');
            } else if (input.trim()) {
                console.log('❓ Unknown command. Type "help" for available commands.');
            }
        });

        // Store readline interface for cleanup
        this.readline = rl;
    }

    // Method to compare original vs transformed documents
    async compareDocuments(originalId) {
        try {
            const sourceDoc = await this.sourceClient
                .db(SOURCE_DB)
                .collection('source_collection_test')
                .findOne({ _id: originalId });

            const maskedDoc = await this.sinkClient
                .db(SINK_DB_MASKED)
                .collection('sink_collection_masked')
                .findOne({ _id: originalId });

            const renamedDoc = await this.sinkClient
                .db(SINK_DB_RENAMED)
                .collection('sink_collection_renamed')
                .findOne({ mongodb_id: originalId });

            console.log('\n🔍 DOCUMENT COMPARISON:');
            console.log('Original:', sourceDoc);
            console.log('Masked Transform:', maskedDoc);
            console.log('Renamed Transform:', renamedDoc);

        } catch (error) {
            console.error('❌ Document comparison failed:', error);
        }
    }
}

// Handle graceful shutdown
process.on('SIGINT', async () => {
    console.log('\n🛑 Received interrupt signal...');
    if (global.monitor) {
        await global.monitor.stop();
    }
    process.exit(0);
});

// Handle script execution
if (require.main === module) {
    const monitor = new TransformMonitor();
    global.monitor = monitor;
    
    monitor.start().catch(error => {
        console.error('💥 Unhandled error:', error);
        process.exit(1);
    });
}

module.exports = TransformMonitor;
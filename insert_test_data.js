#!/usr/bin/env node

// Check for required modules and provide helpful error messages
try {
    const { MongoClient } = require('mongodb');
    const { v4: uuidv4 } = require('uuid');
} catch (error) {
    if (error.code === 'MODULE_NOT_FOUND') {
        console.error('\n❌ Missing required Node.js modules!');
        console.error('\n🔧 To fix this, run the following command:');
        console.error('   npm install mongodb uuid');
        console.error('\n📝 Or install all dependencies from package.json:');
        console.error('   npm install');
        console.error('\n💡 Make sure you have Node.js installed and you\'re in the correct project directory.\n');
        process.exit(1);
    }
    throw error;
}

const { MongoClient } = require('mongodb');
const { v4: uuidv4 } = require('uuid');

// MongoDB connection configuration
const MONGODB_URI = process.env.MONGODB_URI || 'mongodb://localhost:27017';
const DATABASE_NAME = 'mdb_kafka_test';
const COLLECTION_NAME = 'connector_test';

// Insert configuration
const LOCATION = 'Barcelona';
const INSERT_COUNT = parseInt(process.env.INSERT_COUNT) || 1;
const INSERT_INTERVAL = parseInt(process.env.INSERT_INTERVAL) || 0; // milliseconds

async function insertTestData() {
    const client = new MongoClient(MONGODB_URI);
    
    try {
        console.log('Connecting to MongoDB...');
        await client.connect();
        console.log('Connected successfully to MongoDB');
        
        const db = client.db(DATABASE_NAME);
        const collection = db.collection(COLLECTION_NAME);
        
        console.log(`Inserting ${INSERT_COUNT} document(s) into ${DATABASE_NAME}.${COLLECTION_NAME}`);
        
        for (let i = 0; i < INSERT_COUNT; i++) {
            const document = {
                eventID: uuidv4(),
                location: LOCATION,
                timestamp: new Date(),
                insertIndex: i + 1
            };
            
            const result = await collection.insertOne(document);
            console.log(`✓ Inserted document ${i + 1}/${INSERT_COUNT}:`, {
                eventID: document.eventID,
                location: document.location,
                insertedId: result.insertedId
            });
            
            // Wait between inserts if interval is specified
            if (INSERT_INTERVAL > 0 && i < INSERT_COUNT - 1) {
                console.log(`  Waiting ${INSERT_INTERVAL}ms before next insert...`);
                await new Promise(resolve => setTimeout(resolve, INSERT_INTERVAL));
            }
        }
        
        console.log(`\n🎉 Successfully inserted ${INSERT_COUNT} document(s)`);
        
        // Display recent documents
        console.log('\n📋 Recent documents in collection:');
        const recentDocs = await collection
            .find({})
            .sort({ timestamp: -1 })
            .limit(5)
            .toArray();
            
        recentDocs.forEach((doc, index) => {
            console.log(`  ${index + 1}. EventID: ${doc.eventID}, Location: ${doc.location}, Time: ${doc.timestamp?.toISOString()}`);
        });
        
    } catch (error) {
        console.error('❌ Error:', error.message);
        process.exit(1);
    } finally {
        await client.close();
        console.log('\n📤 MongoDB connection closed');
    }
}

// Handle command line arguments
function showHelp() {
    console.log(`
Usage: node insert_test_data.js [options]

Options:
  --help, -h          Show this help message
  --count, -c <num>   Number of documents to insert (default: 1)
  --interval, -i <ms> Milliseconds to wait between inserts (default: 0)
  --uri <uri>         MongoDB connection URI (default: mongodb://localhost:27017)

Environment Variables:
  MONGODB_URI         MongoDB connection URI
  INSERT_COUNT        Number of documents to insert
  INSERT_INTERVAL     Milliseconds between inserts

Examples:
  node insert_test_data.js
  node insert_test_data.js --count 5 --interval 1000
  MONGODB_URI=mongodb://localhost:27018 node insert_test_data.js --count 10
`);
}

// Parse command line arguments
const args = process.argv.slice(2);
for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
        case '--help':
        case '-h':
            showHelp();
            process.exit(0);
            break;
        case '--count':
        case '-c':
            if (i + 1 < args.length) {
                process.env.INSERT_COUNT = args[i + 1];
                i++; // Skip next argument
            }
            break;
        case '--interval':
        case '-i':
            if (i + 1 < args.length) {
                process.env.INSERT_INTERVAL = args[i + 1];
                i++; // Skip next argument
            }
            break;
        case '--uri':
            if (i + 1 < args.length) {
                process.env.MONGODB_URI = args[i + 1];
                i++; // Skip next argument
            }
            break;
    }
}

// Run the script
if (require.main === module) {
    insertTestData().catch(console.error);
}

module.exports = { insertTestData };
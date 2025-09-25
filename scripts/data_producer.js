const { MongoClient } = require('mongodb');

// MongoDB connection settings
const MONGO_URI = 'mongodb://localhost:27017/replset?readPreference=secondary&replicaSet=replset';  
const DATABASE = 'source_db_test';
const COLLECTION = 'source_collection_test';

// Configuration
const BATCH_SIZE = 3;
const BATCH_DELAY = 5000; // 2 seconds between batches
const TOTAL_BATCHES = 2;

class DataProducer {
    constructor() {
        this.client = new MongoClient(MONGO_URI);
        this.db = null;
        this.collection = null;
        this.host = null;
    }

    async connect() {
        try {
            await this.client.connect();
            this.db = this.client.db(DATABASE);
            this.collection = this.db.collection(COLLECTION);
            console.log(`Connected to MongoDB: ${DATABASE}.${COLLECTION}`);
        } catch (error) {
            console.error('Failed to connect to MongoDB:', error);
            throw error;
        }
    }

    generateTestDocument(index) {
        const emails = [
            'john.doe@example.com',
            'jane.smith@company.org',
            'admin@testsite.net',
            'user123@domain.co.uk',
            'contact@business.com'
        ];

        return {
            _id: `user_${index}_${Date.now()}`,
            userId: `USR${String(index).padStart(4, '0')}`,
            name: `Test User ${index}`,
            email: emails[index % emails.length],
            contact_email: `contact_${index}@example.com`,
            user_email: `user${index}@test.org`,
            profile: {
                age: 20 + (index % 50),
                city: ['New York', 'London', 'Tokyo', 'Paris', 'Berlin'][index % 5],
                preferences: {
                    newsletter: index % 2 === 0,
                    notifications: index % 3 === 0
                }
            },
            metadata: {
                created_at: new Date(),
                batch_id: Math.floor(index / BATCH_SIZE) + 1,
                source: 'data_producer_script'
            },
            // Add some variety in document structure
            ...(index % 3 === 0 && { 
                optional_field: `Optional data for record ${index}`,
                tags: [`tag${index % 5}`, `category${index % 3}`]
            })
        };
    }

    async insertBatch(batchNumber) {
        const documents = [];
        const startIndex = (batchNumber - 1) * BATCH_SIZE;
        
        for (let i = 0; i < BATCH_SIZE; i++) {
            documents.push(this.generateTestDocument(startIndex + i));
        }

        try {
            const result = await this.collection.insertMany(documents);
            console.log(`✅ Batch ${batchNumber} inserted: ${result.insertedCount} documents`);
            
            // Log sample document for verification
            if (batchNumber === 1) {
                console.log('📄 Sample document structure:');
                console.log(JSON.stringify(documents[0], null, 2));
            }
            
            return result;
        } catch (error) {
            console.error(`❌ Failed to insert batch ${batchNumber}:`, error);
            throw error;
        }
    }

    async run() {
        console.log('🚀 Starting Data Producer for SMT Demo');
        console.log(`Configuration:
                    - Database: ${DATABASE}
                    - Collection: ${COLLECTION}
                    - Batch Size: ${BATCH_SIZE}
                    - Total Batches: ${TOTAL_BATCHES}
                    - Delay Between Batches: ${BATCH_DELAY}ms
        `);

        try {
            await this.connect();

            for (let batch = 1; batch <= TOTAL_BATCHES; batch++) {
                console.log(`\n📦 Processing batch ${batch}/${TOTAL_BATCHES}...`);
                await this.insertBatch(batch);
                
                if (batch < TOTAL_BATCHES) {
                    console.log(`⏳ Waiting ${BATCH_DELAY}ms before next batch...`);
                    await new Promise(resolve => setTimeout(resolve, BATCH_DELAY));
                }
            }

            // Final statistics
            const totalCount = await this.collection.countDocuments({
                'metadata.source': 'data_producer_script'
            });
            
            console.log(`\n✅ Data production completed!`);
            console.log(`📊 Total documents inserted: ${totalCount}`);
            console.log(`\n💡 Next steps:
                    1. Deploy a sink connector with SMT transforms
                    2. Monitor the transformation effects
                    3. Use scripts/monitor_transforms.js to observe changes`);

        } catch (error) {
            console.error('❌ Data production failed:', error);
        } finally {
            await this.client.close();
            console.log('🔌 MongoDB connection closed');
        }

    }
}

// Handle script execution
if (require.main === module) {
    const producer = new DataProducer();
    producer.run().catch(error => {
        console.error('💥 Unhandled error:', error);
        process.exit(1);
    });
}

module.exports = DataProducer;
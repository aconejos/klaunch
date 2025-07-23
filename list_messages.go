package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func list_messages() error {
	// Get available topics
	availableTopics, err := list_topics()
	if err != nil {
		fmt.Println("Error getting topics:", err)
		return err
	}

	var topicName string
	if availableTopics == nil || len(availableTopics) == 0 {
		fmt.Println("No topics found. Please enter a topic name manually:")
		fmt.Print("Topic name: ")
		fmt.Scanln(&topicName)
	} else {
		// Display available topics with menu
		fmt.Println("\nAvailable topics:")
		for i, topic := range availableTopics {
			fmt.Printf("%d. %s\n", i+1, topic)
		}
		fmt.Printf("\nSelect a topic (1-%d): ", len(availableTopics))

		var choice string
		fmt.Scanln(&choice)

		choiceNum, err := strconv.Atoi(choice)
		if err != nil || choiceNum < 1 || choiceNum > len(availableTopics) {
			return fmt.Errorf("invalid choice: please select a number between 1 and %d", len(availableTopics))
		} else {
			// Selected from available topics
			topicName = availableTopics[choiceNum-1]
		}
	}

	// Validate topic name
	if topicName == "" {
		return fmt.Errorf("no topic name provided")
	}

	fmt.Printf("Selected topic: %s\n", topicName)

	// define the request
	brokers := "localhost:9091,localhost:9092,localhost:9093"
	topics := []string{topicName}
	group := "consumer-cluster-group"
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("Connecting to Kafka brokers: %s\n", brokers)
	fmt.Printf("Consumer group: %s\n", group)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        brokers,
		"group.id":                 group,
		"session.timeout.ms":       30000,  // Increased from 6000 to 30000
		"heartbeat.interval.ms":    10000,  // Added heartbeat interval
		"max.poll.interval.ms":     300000, // Added max poll interval
		"receive.message.max.bytes": 2147483647,
		"security.protocol":        "PLAINTEXT",
		"api.version.request":      true,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  5000,
	})

	if err != nil {
		fmt.Printf("Failed to create consumer: %v\n", err)
		return err
	}

	fmt.Printf("✓ Created Consumer successfully\n")

	fmt.Printf("Subscribing to topics: %v\n", topics)
	err = c.SubscribeTopics(topics, nil)
	if err != nil {
		fmt.Printf("Error subscribing to topics: %v\n", err)
		c.Close()
		return err
	}

	fmt.Printf("✓ Successfully subscribed to topics\n")
	fmt.Println("Listening for messages... (Press Ctrl+C to stop)")
	fmt.Println("---")

	run := true

	for run == true {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				parts := make([]kafka.TopicPartition, len(e.Partitions))
				for i, tp := range e.Partitions {
					tp.Offset = kafka.OffsetTail(10) // Start from last 10 messages
					parts[i] = tp
				}
				fmt.Printf("✓ Assigned partitions: %v\n", parts)
				c.Assign(parts)
			case *kafka.Message:
				fmt.Printf("\n📨 Message received:\n")
				fmt.Printf("Topic: %s\n", *e.TopicPartition.Topic)
				fmt.Printf("Partition: %d\n", e.TopicPartition.Partition)
				fmt.Printf("Offset: %d\n", e.TopicPartition.Offset)
				if e.Key != nil {
					fmt.Printf("Key: %s\n", string(e.Key))
				}
				fmt.Printf("Value: %s\n", string(e.Value))
				if e.Timestamp.IsZero() == false {
					fmt.Printf("Timestamp: %v\n", e.Timestamp)
				}
				fmt.Println("---")
			case kafka.PartitionEOF:
				fmt.Printf("📍 Reached end of partition %v\n", e)
			case kafka.Error:
				if e.Code() == kafka.ErrTimedOut {
					fmt.Printf("⏱️  No messages received (timeout)\n")
				} else {
					fmt.Fprintf(os.Stderr, "❌ Kafka Error: %v\n", e)
					run = false
				}
			case kafka.RevokedPartitions:
				fmt.Printf("🔄 Partitions revoked: %v\n", e)
			default:
				// Ignore other events like OffsetsCommitted
			}
		}
	}

	fmt.Printf("Closing consumer\n")
	c.Close()
	return nil
}

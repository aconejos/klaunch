package main

import (
	"fmt"
	"os/exec"
	"bufio"

	// "os/signal"
	//
	//	"syscall"
	//
	//	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func list_messages() error {
	// show the list of topics
	list_topics()

	// request the topic name
	fmt.Println("Enter topic name: ")
	var topicName string
	fmt.Scanln(&topicName)
	//fmt.Println("docker exec kafka-connect kafka-console-consumer --topic disable.db_name.coll_name --from-beginning --bootstrap-server=kafka2:19092,kafka3:19093,kafka1:19091")

	listCmd := exec.Command("docker", "exec", "kafka-connect", "kafka-console-consumer", "--topic", topicName, "--from-beginning", "--bootstrap-server=kafka2:19092,kafka3:19093,kafka1:19091")
    
	fmt.Println("To list the messages plese execute this command: ", listCmd)


	out, err := listCmd.StdoutPipe()
    if err != nil {
        return err
    }
	
	errOut, _ := listCmd.StderrPipe() // Capture stderr
    if errOut != nil {
        fmt.Println("Error running this command on polling:", errOut)
        return nil
    }

	err = listCmd.Start()
    if err != nil {
        return err
    }

	fmt.Println("the command ", listCmd)
    scanner := bufio.NewScanner(out)
	fmt.Println("the command out ", scanner.Scan())

    go func() {
		fmt.Println("Arranca...")
        for scanner.Scan() {
			fmt.Println("scanneando...")
            fmt.Println(scanner.Text())
        }
    }()

    err = listCmd.Wait()
    if err != nil {
        fmt.Println("Command finished with error:", err)
    }

	return nil

	//// define the request
	//broker := "host.docker.internal"
	//topics := []string{topicName}
	//group := "connect-cluster-group"
	//sigchan := make(chan os.Signal, 1)
	//signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	//
	//c, err := kafka.NewConsumer(&kafka.ConfigMap{
	//	"bootstrap.servers":               broker,
	//	"group.id":                        group,
	//	"go.application.rebalance.enable": true, // delegate Assign() responsibility to app
	//	"session.timeout.ms":              6000,
	//	"receive.message.max.bytes":       2147483647,
	//	"security.protocol":               "PLAINTEXT",
	//	"api.version.request":             1,
	//	"default.topic.config":            kafka.ConfigMap{"auto.offset.reset": "earliest"}})
	//
	//if err != nil {
	//	fmt.Printf("Failed to create consumer")
	//	return err
	//}
	//
	//fmt.Printf("Created Consumer %v\n", c)
	//
	//err = c.SubscribeTopics(topics, nil)
	//if err != nil {
	//	fmt.Printf("Error reading topics: ")
	//	return err
	//}
	//
	//run := true
	//
	//for run == true {
	//	select {
	//	case sig := <-sigchan:
	//		fmt.Printf("Caught signal %v: terminating\n", sig)
	//		run = false
	//	default:
	//		ev := c.Poll(100)
	//		if ev == nil {
	//			continue
	//		}
	//
	//		switch e := ev.(type) {
	//		case kafka.AssignedPartitions:
	//			parts := make([]kafka.TopicPartition,
	//				len(e.Partitions))
	//			for i, tp := range e.Partitions {
	//				tp.Offset = kafka.OffsetTail(5) // Set start offset to 5 messages from end of partition
	//				parts[i] = tp
	//			}
	//			fmt.Printf("Assign %v\n", parts)
	//			c.Assign(parts)
	//		case *kafka.Message:
	//			fmt.Printf("%% Message on %s:\n%s\n",
	//				e.TopicPartition, string(e.Value))
	//		case kafka.PartitionEOF:
	//			fmt.Printf("%% Reached %v\n", e)
	//		case kafka.Error:
	//			fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
	//			run = false
	//		default:
	//			fmt.Printf("Ignored %v\n", e)
	//		}
	//	}
	//}
	//
	//fmt.Printf("Closing consumer\n")
	//c.Close()
	//return nil
}

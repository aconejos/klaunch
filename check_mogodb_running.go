package main

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func check_mongodb_running() error {
	// Replica set addresses
	replSetAddresses := []string{"127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000", "127.0.0.1:27018/?directConnection=true&serverSelectionTimeoutMS=2000", "127.0.0.1:27019/?directConnection=true&serverSelectionTimeoutMS=2000"}

	for _, addr := range replSetAddresses {
		// Attempt to connect to MongoDB
		clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", addr))
		client, _ := mongo.Connect(context.Background(), clientOptions)
		err := client.Ping(context.Background(), nil)
		if err != nil {
			return err
		}
		var isMasterResult bson.M
		db := client.Database("admin") // Change "admin" to your target database name

		err = db.RunCommand(context.Background(), bson.D{{Key: "ismaster", Value: true}}).Decode(&isMasterResult)
		// if isMasterResult is true then print the connections string
		if isMasterResult["ismaster"] == true {
			fmt.Printf("Connected to MongoDB at %s\n", addr)
			// define the script to runn on the primary node
			script := `
			var cfg = rs.conf();
			cfg.members[0].host = "host.docker.internal:27017";
			cfg.members[1].host = "host.docker.internal:27018";
			cfg.members[2].host = "host.docker.internal:27019";
			rs.reconfig(cfg);
			`

			// find the port number
			port := strings.Trim(regexp.MustCompile(`:\d+`).FindStringSubmatch(addr)[0], ":")


			// Execute the script
			cmd := exec.Command("mongosh", "--port", port, "--quiet","--eval", script)

			//print the command 
			//fmt.Println(cmd.Args)

			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Output: %s\n", string(output))
				return err
			}

			client.Disconnect(context.Background())
			return nil
		}
		if err != nil {
			client.Disconnect(context.Background())
			return fmt.Errorf("failed to run ismaster command: %v", err)
		}
	}

	fmt.Println("Successfully connected to MongoDB")
	return nil
}

package main

import (
	"fmt"
	"kvstore_lab2/common"
	"log"
	"net/rpc"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("  go run client.go put <key> <value>")
		fmt.Println("  go run client.go get <key>")
		return
	}

	client, err := rpc.Dial("tcp", "localhost:9000")
	if err != nil {
		log.Fatal("Connection to master failed:", err)
	}

	command := os.Args[1]

	switch command {
	case "put":
		if len(os.Args) != 4 {
			fmt.Println("Usage: go run client.go put <key> <value>")
			return
		}
		args := common.Args{Key: os.Args[2], Value: os.Args[3]}
		var reply common.Reply
		err = client.Call("Master.Put", args, &reply)
		if err != nil {
			log.Fatal("Put error:", err)
		}
		fmt.Println("Put result:", reply.Status)

	case "get":
		if len(os.Args) != 3 {
			fmt.Println("Usage: go run client.go get <key>")
			return
		}
		args := common.Args{Key: os.Args[2]}
		var reply common.Reply
		err = client.Call("Master.Get", args, &reply)
		if err != nil {
			log.Fatal("Get error:", err)
		}
		fmt.Printf("Get result: %s (status: %s)\n", reply.Value, reply.Status)

	default:
		fmt.Println("Unknown command:", command)
	}
}

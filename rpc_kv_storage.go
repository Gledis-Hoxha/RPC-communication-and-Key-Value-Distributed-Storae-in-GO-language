package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"sync"
)

type KeyValueStore struct {
	mu    sync.Mutex
	Store map[string]string
}

type PutArgs struct {
	Key   string
	Value string
}

func (kvs *KeyValueStore) Put(args *PutArgs, reply *bool) error {
	kvs.mu.Lock()
	defer kvs.mu.Unlock()
	kvs.Store[args.Key] = args.Value
	saveToFile(kvs.Store)
	*reply = true
	return nil
}

func (kvs *KeyValueStore) Get(key *string, reply *string) error {
	kvs.mu.Lock()
	defer kvs.mu.Unlock()
	*reply = kvs.Store[*key]
	return nil
}

func (kvs *KeyValueStore) Delete(key *string, reply *bool) error {
	kvs.mu.Lock()
	defer kvs.mu.Unlock()
	delete(kvs.Store, *key)
	saveToFile(kvs.Store)
	*reply = true
	return nil
}

func (kvs *KeyValueStore) List(args *struct{}, reply *map[string]string) error {
	kvs.mu.Lock()
	defer kvs.mu.Unlock()
	*reply = kvs.Store
	return nil
}

func saveToFile(store map[string]string) {
	file, _ := json.MarshalIndent(store, "", "  ")
	_ = os.WriteFile("storage.json", file, 0644)
}

func loadFromFile() map[string]string {
	file, err := os.ReadFile("storage.json")
	if err != nil {
		return make(map[string]string)
	}
	var store map[string]string
	_ = json.Unmarshal(file, &store)
	return store
}

func startServer() {
	store := loadFromFile()
	kvs := &KeyValueStore{Store: store}

	rpc.Register(kvs)
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("Server error:", err)
		return
	}
	fmt.Println("Server running on :1234... (Ctrl+C to stop)")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Client connection error:", err)
			continue
		}
		fmt.Printf("New client connected: %s\n", conn.RemoteAddr())
		go rpc.ServeConn(conn)
	}
}

func startClient() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run rpc_kv_storage.go [put/get/delete/list] [key] [value]")
		fmt.Println("Examples:")
		fmt.Println("  go run rpc_kv_storage.go put name Alice")
		fmt.Println("  go run rpc_kv_storage.go get name")
		fmt.Println("  go run rpc_kv_storage.go delete name")
		fmt.Println("  go run rpc_kv_storage.go list")
		return
	}

	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		fmt.Println("Is the server running? Start it with: go run rpc_kv_storage.go server")
		return
	}
	defer client.Close()

	cmd := os.Args[1]

	switch cmd {
	case "put":
		if len(os.Args) != 4 {
			fmt.Println("Usage: put [key] [value]")
			return
		}
		key := os.Args[2]
		value := os.Args[3]
		var reply bool
		err = client.Call("KeyValueStore.Put", &PutArgs{Key: key, Value: value}, &reply)
		if err != nil {
			fmt.Println("Put error:", err)
		} else {
			fmt.Printf("Saved: %s -> %s\n", key, value)
		}
	case "get":
		if len(os.Args) != 3 {
			fmt.Println("Usage: get [key]")
			return
		}
		key := os.Args[2]
		var reply string
		err = client.Call("KeyValueStore.Get", &key, &reply)
		if err != nil {
			fmt.Println("Get error:", err)
		} else {
			if reply == "" {
				fmt.Printf("Key '%s' not found\n", key)
			} else {
				fmt.Printf("Value for '%s': %s\n", key, reply)
			}
		}
	case "delete":
		if len(os.Args) != 3 {
			fmt.Println("Usage: delete [key]")
			return
		}
		key := os.Args[2]
		var reply bool
		err = client.Call("KeyValueStore.Delete", &key, &reply)
		if err != nil {
			fmt.Println("Delete error:", err)
		} else {
			fmt.Printf("Deleted key: %s\n", key)
		}
	case "list":
		var reply map[string]string
		err = client.Call("KeyValueStore.List", &struct{}{}, &reply)
		if err != nil {
			fmt.Println("List error:", err)
		} else {
			if len(reply) == 0 {
				fmt.Println("No data stored")
			} else {
				fmt.Println("Stored data:")
				for k, v := range reply {
					fmt.Printf("  %s -> %s\n", k, v)
				}
			}
		}
	default:
		fmt.Println("Unknown command. Use 'put', 'get', 'delete', or 'list'.")
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		startServer()
	} else {
		startClient()
	}
}

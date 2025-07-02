package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"sync"

	"kvstore_lab2/common"
)

type KVSlave struct {
	mu    sync.Mutex
	store map[string]string
}

func (s *KVSlave) Put(args *common.Args, reply *common.Reply) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[args.Key] = args.Value
	reply.Status = "OK"
	return nil
}

func (s *KVSlave) Get(args *common.Args, reply *common.Reply) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.store[args.Key]
	if !ok {
		reply.Status = "NOT_FOUND"
		return nil
	}
	reply.Value = val
	reply.Status = "OK"
	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run slave.go <port> <id>")
		return
	}

	slave := &KVSlave{store: make(map[string]string)}
	rpc.Register(slave)

	listener, err := net.Listen("tcp", ":"+os.Args[1])
	if err != nil {
		fmt.Println("Listen error:", err)
		return
	}
	fmt.Printf("Slave %s running on port %s\n", os.Args[2], os.Args[1])
	rpc.Accept(listener)
}

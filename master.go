package main

import (
	"fmt"
	"hash/fnv"
	"kvstore_lab2/common"
	"net"
	"net/rpc"
	"sort"
	"time"
)

type SlaveInfo struct {
	ID   int
	Addr string
}

type Master struct {
	Slaves []SlaveInfo
}

// Hash function
func hashKey(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() % 360)
}

func (m *Master) getReplicaSlaves(key string) []SlaveInfo {
	h := hashKey(key)
	sort.Slice(m.Slaves, func(i, j int) bool { return m.Slaves[i].ID < m.Slaves[j].ID })

	var result []SlaveInfo
	for _, s := range m.Slaves {
		if s.ID >= h {
			result = append(result, s)
		}
	}
	result = append(result, m.Slaves...)

	replicas := make([]SlaveInfo, 0, 3)
	seen := map[int]bool{}
	for _, s := range result {
		if !seen[s.ID] {
			replicas = append(replicas, s)
			seen[s.ID] = true
			if len(replicas) == 3 {
				break
			}
		}
	}

	return replicas
}

// Put për klientin
func (m *Master) Put(args common.Args, reply *common.Reply) error {
	replicas := m.getReplicaSlaves(args.Key)
	success := false
	done := make(chan bool, len(replicas))

	for _, slave := range replicas {
		go func(sl common.Args, addr string) {
			client, err := rpc.Dial("tcp", addr)
			if err == nil {
				defer client.Close()
				var slaveReply common.Reply
				err = client.Call("KVSlave.Put", &sl, &slaveReply)
				if err == nil && slaveReply.Status == "OK" {
					done <- true
					return
				}
			}
			done <- false
		}(args, slave.Addr)
	}

	timeout := time.After(2 * time.Second)
	for i := 0; i < len(replicas); i++ {
		select {
		case ok := <-done:
			if ok {
				success = true
				break
			}
		case <-timeout:
			break
		}
	}

	if success {
		reply.Status = "OK"
	} else {
		reply.Status = "ERROR"
	}
	return nil
}

// Get për klientin
func (m *Master) Get(args common.Args, reply *common.Reply) error {
	replicas := m.getReplicaSlaves(args.Key)

	for _, slave := range replicas {
		client, err := rpc.Dial("tcp", slave.Addr)
		if err == nil {
			defer client.Close()
			var slaveReply common.Reply
			err = client.Call("KVSlave.Get", &args, &slaveReply)
			if err == nil && slaveReply.Status == "OK" {
				reply.Value = slaveReply.Value
				reply.Status = "OK"
				return nil
			}
		}
	}

	reply.Status = "NOT_FOUND"
	return nil
}

func main() {
	master := &Master{
		Slaves: []SlaveInfo{
			{ID: 10, Addr: "localhost:8001"},
			{ID: 70, Addr: "localhost:8002"},
			{ID: 130, Addr: "localhost:8003"},
			{ID: 200, Addr: "localhost:8004"},
			{ID: 310, Addr: "localhost:8005"},
		},
	}

	rpc.RegisterName("Master", master)

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println("Master listen error:", err)
		return
	}

	fmt.Println("Master RPC server running on port 9000")
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}
}

Distributed Key-Value Storage System (Go + RPC)
Overview
This project implements a distributed key-value storage system using Go and RPC for communication. It consists of a Master-Slave architecture to provide a scalable and highly available storage service.

Master node: Receives all client requests (put(key, value), get(key)), determines the appropriate Slave nodes to store or fetch the key-value pairs, manages replication, and handles client responses.

Slave nodes: Store key-value data persistently on disk and serve read/write requests from the Master node.

Features
RPC-based communication between client, Master, and Slave nodes over TCP.

Persistent storage of key-value pairs on each Slave node (data saved in files).

Master-Slave architecture with:

1 Master node managing 5 Slave nodes.

Key distribution and replication based on consistent hashing (Chord-like ring).

Each key-value replicated on 3 consecutive Slave nodes for fault tolerance.

Fault tolerance: Master waits for at least 1 confirmation from the 3 replicas on put. For get, Master tries the 3 replicas in order until one responds or times out.

Components
kvClient: CLI client to send put and get commands to Master.

Usage:

kvClient put <key> <value>

kvClient get <key>

Master server: Coordinates storage and retrieval, manages replication, and handles client requests.

Slave servers (5 nodes): Store key-value data persistently and respond to Master's requests.

How it works
put(key, value):

Client sends request to Master.

Master calculates which 3 Slave nodes (consecutive in ring) should store the key.

Master sends write request to all 3 Slaves concurrently.

On receiving at least one successful confirmation, Master replies success to client.

If no confirmation within timeout, Master replies error.

get(key):

Client sends request to Master.

Master finds the 3 replicas holding the key.

Master queries one Slave randomly, waits for response.

If no response, tries next Slave in ring.

If none respond, Master returns error.

Persistence:

Each Slave stores its key-value map in a local file on disk for persistence.

Running the project
Start the 5 Slave servers (each with their own IP and port).

Start the Master server, configured with the Slave nodes’ IPs and IDs.

Use kvClient to send put and get commands to the Master.

Dependencies
Go language (1.18+ recommended)

Standard Go RPC package over TCP

Notes
Consistent hashing and replication logic follows a simplified Chord protocol.

Timeouts and retries are implemented to ensure availability and fault tolerance.

Data stored on disk by Slaves enables persistence across restarts.


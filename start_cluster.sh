#!/bin/bash

# Clean up old data to ensure a fresh election
echo "Cleaning up old Raft data..."
rm -rf ./raft-data
mkdir -p ./raft-data/node-1
mkdir -p ./raft-data/node-2
mkdir -p ./raft-data/node-3

# Build the binary
echo "Building Gopher-Guard..."
go build -o gopher-guard ./cmd/server/main.go

# Start Node 1 (Bootstrap Node / Leader)
echo "Starting Node 1..."
PORT=50051 ADMIN_PORT=8080 NODE_ID=node-1 BOOTSTRAP=true RAFT_BIND=127.0.0.1:7000 RAFT_DIR=./raft-data/node-1 ./gopher-guard &
NODE1_PID=$!

sleep 2 # Give Node 1 a moment to write to disk and become Leader

# Start Node 2
echo "Starting Node 2..."
PORT=50052 ADMIN_PORT=8081 NODE_ID=node-2 RAFT_BIND=127.0.0.1:7001 RAFT_DIR=./raft-data/node-2 ./gopher-guard &
NODE2_PID=$!

# Start Node 3
echo "Starting Node 3..."
PORT=50053 ADMIN_PORT=8082 NODE_ID=node-3 RAFT_BIND=127.0.0.1:7002 RAFT_DIR=./raft-data/node-3 ./gopher-guard &
NODE3_PID=$!

trap "echo -e '\nShutting down cluster...'; kill -SIGTERM $NODE1_PID $NODE2_PID $NODE3_PID; wait; exit" SIGINT

echo "======================================================"
echo "🛡️ Gopher-Guard Cluster Running! Press Ctrl+C to stop."
echo "======================================================"

wait
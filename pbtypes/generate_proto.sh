#!/bin/bash

# Script to generate protobuf code from monitor.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative monitor/monitor.proto

echo "Protobuf generation completed."
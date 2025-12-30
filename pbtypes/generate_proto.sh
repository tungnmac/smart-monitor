#!/bin/bash

# Script to generate protobuf code from all .proto files in subdirectories

PROTO_PATH="../third_party/googleapis"

# Collect all proto files
PROTO_FILES=$(find . -name "*.proto" -type f | grep -v third_party)

echo "Generating individual protobuf files..."

# Generate for each proto file individually
for proto_file in $PROTO_FILES; do
    dir=$(dirname "$proto_file")
    base=$(basename "$proto_file" .proto)

    echo "Generating for $proto_file"

    # Generate Go and gRPC code
    protoc -I. -I$PROTO_PATH --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative "$proto_file"

    # Generate grpc-gateway code
    protoc -I. -I$PROTO_PATH --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative "$proto_file"

    # Generate OpenAPI/Swagger individually
    protoc -I. -I$PROTO_PATH --openapiv2_out="$dir" --openapiv2_opt=allow_merge=true,merge_file_name="$base" "$proto_file"
done

echo "Generating combined swagger..."

# Generate combined OpenAPI/Swagger
protoc -I. -I$PROTO_PATH --openapiv2_out=. --openapiv2_opt=allow_merge=true,merge_file_name=combined $PROTO_FILES

echo "Protobuf generation completed."
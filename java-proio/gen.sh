#!/bin/bash

# Generate necessary protobuf messages
protoc --proto_path=proio/proto=../proto --java_out=src/main/java ../proto/proio.proto

# Generate protobuf messages for common data models
for proto in $(find ../model -iname "*.proto"); do
    protoc --proto_path=proio/model=../model --java_out=src/main/java $proto
done

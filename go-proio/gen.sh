#!/bin/bash

# Generate necessary protobuf messages
protoc --proto_path=proio/proto=../proto --gofast_out=$GOPATH/src ../proto/proio.proto

# Generate protobuf messages for common data models
for proto in $(find ../model -iname "*.proto"); do
    protoc --proto_path=proio/model=../model --gofast_out=$GOPATH/src $proto
done

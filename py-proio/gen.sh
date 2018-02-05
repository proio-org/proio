#!/bin/bash

# Generate necessary protobuf messages
protoc --proto_path=proio/proto=../proto --python_out=. ../proto/proio.proto

# Generate protobuf messages for common data models
for proto in $(find ../model -iname "*.proto"); do
    protoc --proto_path=proio/model=../model --python_out=. $proto
done

touch proio/model/__init__.py

for gen_file in $(ls proio/model/*_pb2.py); do
    mod_dir=${gen_file%_pb2.py}
    mkdir $mod_dir &> /dev/null
    mv $gen_file $mod_dir/
    gen_mod=${gen_file##*/}
    echo "from .${gen_mod%.py} import *" > $mod_dir/__init__.py
done

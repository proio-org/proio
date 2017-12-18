# Generate necessary protobuf messages
protoc --proto_path=proio/proto=../proto --python_out=. ../proto/proio.proto

# Generate protobuf messages for common data models
for proto in $(find ../model -iname "*.proto"); do
    protoc --proto_path=proio/model=../model --python_out=. $proto
done

for gen_file in $(find proio/model -iname "*_pb2.py"); do
    mkdir ${gen_file%_pb2.py}
    gen_dir=${gen_file%_pb2.py}
    mv $gen_file $gen_dir/
    gen_file=${gen_file##*/}
    echo "from .${gen_file%.py} import *" > $gen_dir/__init__.py
done

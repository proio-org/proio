.PHONY: all

PROTO := eicio.proto

GO_TARGET := go-eicio/eicio.pb.go
CPP_TARGET := cpp-eicio/eicio.pb.h

TARGETS := $(GO_TARGET) $(CPP_TARGET)

all: $(TARGETS)

$(GO_TARGET): $(PROTO)
	protoc --gofast_out=$(@D) $<
	sed -i '/\/\*/,/\*\//d' $@

$(CPP_TARGET): $(PROTO)
	protoc --cpp_out=$(@D) $<

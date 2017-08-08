.PHONY: all

PROTO := eicio.proto

GO_PATH := go-eicio
CPP_PATH := cpp-eicio

GO_TARGET := $(GO_PATH)/eicio.pb.go
CPP_TARGET := $(CPP_PATH)/eicio.pb.h

TARGETS := $(GO_TARGET) \
			$(CPP_TARGET)

all: $(TARGETS)

$(GO_TARGET): $(PROTO)
	protoc --gofast_out=$(GO_PATH) $<

$(CPP_TARGET): $(PROTO)
	protoc --cpp_out=$(CPP_PATH) $<

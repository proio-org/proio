PROTO := eicio.proto

GO_TARGET := go-eicio/eicio.pb.go
CPP_TARGET := cpp-eicio/src/eicio.pb.h cpp-eicio/src/eicio.pb.cc

TARGETS := $(GO_TARGET) $(CPP_TARGET)

.PHONY: all clean

all: $(TARGETS)

clean: 
	rm -f $(GO_TARGET) $(CPP_TARGET)

$(GO_TARGET): $(PROTO) go-eicio/genExtraMsgFuncs.sh
	protoc --gofast_out=$(@D) $<
	sed -i '/\/\*/,/\*\//d' $@
	. go-eicio/genExtraMsgFuncs.sh $< $@

$(CPP_TARGET): $(PROTO)
	protoc --cpp_out=$(@D) $<
	cd $(@D) && clang-format -i -style=file $(notdir $(CPP_TARGET))

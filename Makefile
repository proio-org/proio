PROTO := eicio.proto

GO_TARGET := go-eicio/model/eicio.pb.go
CPP_TARGET := cpp-eicio/src/eicio/eicio.pb.h cpp-eicio/src/eicio/eicio.pb.cc
PYTHON_TARGET := py-eicio/eicio/eicio_pb2.py

TARGETS := $(GO_TARGET) $(CPP_TARGET) $(PYTHON_TARGET)

.PHONY: all clean

all: $(TARGETS)

clean: 
	rm -f $(TARGETS)

# call to genExtraMsgFuncs may be removed later.  This is to avoid expensive
# reflection, but there may be another way I'm not seeing right now.
$(GO_TARGET): $(PROTO) go-eicio/genExtraMsgFuncs.sh
	protoc --gofast_out=$(@D) $<
	sed -i '/\/\*/,/\*\//d' $@
	. go-eicio/genExtraMsgFuncs.sh $< $@

$(CPP_TARGET): $(PROTO)
	protoc --cpp_out=$(@D) $<
	cd $(@D) && clang-format -i -style=file $(notdir $(CPP_TARGET))

$(PYTHON_TARGET): $(PROTO)
	protoc --python_out=$(@D) $<

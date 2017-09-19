PROTO := eicio.proto

BUILD_IMAGE := eicio-build.img
COMMAND_PREFIX := singularity exec -C -H $(PWD) $(BUILD_IMAGE)

GO_TARGET := go-eicio/model/eicio.pb.go
CPP_TARGET := cpp-eicio/src/eicio/eicio.pb.h cpp-eicio/src/eicio/eicio.pb.cc
PYTHON_TARGET := py-eicio/eicio/model/eicio_pb2.py

TARGETS := $(GO_TARGET) $(CPP_TARGET) $(PYTHON_TARGET)

.PHONY: all clean

all: $(TARGETS)

clean: 
	rm -f $(TARGETS)

$(BUILD_IMAGE):
	SINGULARITY_CACHEDIR=/tmp/singularity-cache singularity pull -n $@ docker://dbcooper/eicio-build

# call to genExtraMsgFuncs may be removed later.  This is to avoid expensive
# reflection, but there may be another way I'm not seeing right now.
$(GO_TARGET): $(PROTO) go-eicio/genExtraMsgFuncs.sh $(BUILD_IMAGE)
	$(COMMAND_PREFIX) protoc --gofast_out=$(@D) $<
	$(COMMAND_PREFIX) bash -c ". go-eicio/genExtraMsgFuncs.sh $< $@"

$(CPP_TARGET): $(PROTO) $(BUILD_IMAGE)
	$(COMMAND_PREFIX) protoc --cpp_out=$(@D) $<
	$(COMMAND_PREFIX) bash -c "cd cpp-eicio/src && clang-format -i -style=file $(subst cpp-eicio/src/,,$(CPP_TARGET))"

$(PYTHON_TARGET): $(PROTO) $(BUILD_IMAGE)
	$(COMMAND_PREFIX) protoc --python_out=$(@D) $<

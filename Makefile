PROTO := proio.proto

BUILD_IMAGE := proio-gen.img
COMMAND_PREFIX := singularity exec -C -H $(PWD) $(BUILD_IMAGE)

GO_TARGET := go-proio/model/proio.pb.go
CPP_TARGET := cpp-proio/src/proio/proio.pb.h cpp-proio/src/proio/proio.pb.cc
PYTHON_TARGET := py-proio/proio/model/proio_pb2.py
JAVA_TARGET := java-proio/src/main/java/proio/Model.java

TARGETS := $(GO_TARGET) $(CPP_TARGET) $(PYTHON_TARGET) $(JAVA_TARGET)

.PHONY: all clean

all: $(TARGETS)

clean: 
	rm -f $(TARGETS)

$(BUILD_IMAGE):
	SINGULARITY_CACHEDIR=/tmp/singularity-cache singularity pull -n $@ docker://dbcooper/proio-gen

# call to genExtraMsgFuncs may be removed later.  This is to avoid expensive
# reflection, but there may be another way I'm not seeing right now.
$(GO_TARGET): $(PROTO) go-proio/genExtraMsgFuncs.sh $(BUILD_IMAGE)
	$(COMMAND_PREFIX) protoc --gofast_out=$(@D) $<
	$(COMMAND_PREFIX) bash -c ". go-proio/genExtraMsgFuncs.sh $< $@"

$(CPP_TARGET): $(PROTO) $(BUILD_IMAGE)
	$(COMMAND_PREFIX) protoc --cpp_out=$(@D) $<
	$(COMMAND_PREFIX) bash -c "cd cpp-proio/src && clang-format -i -style=file $(subst cpp-proio/src/,,$(CPP_TARGET))"

$(PYTHON_TARGET): $(PROTO) $(BUILD_IMAGE)
	$(COMMAND_PREFIX) protoc --python_out=$(@D) $<

$(JAVA_TARGET): $(PROTO) $(BUILD_IMAGE)
	$(COMMAND_PREFIX) protoc --java_out=$(patsubst %/proio/Model.java,%,$(JAVA_TARGET)) $<

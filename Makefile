SOURCE := $(shell find model -name "*.proto") $(shell find proto -name "*.proto")
GO_TARGETS := $(addprefix go-proio/,$(SOURCE:.proto=.pb.go))
CPP_TARGETS := $(addprefix cpp-proio/src/proio/,$(SOURCE:.proto=.pb.cc))
CPP_HEADERS := $(addprefix cpp-proio/src/proio/,$(SOURCE:.proto=.pb.h))
PYTHON_TARGETS := $(addprefix py-proio/proio/,$(SOURCE:.proto=_pb2.py))
JAVA_TARGETS := $(addprefix java-proio/src/main/java/,$(foreach source,$(SOURCE),$(shell grep java_package $(source) | sed 's/option\s*java_package\s*=\s*"\(.*\)";/\1/')/$(shell grep java_outer_classname $(source) | sed 's/option\s*java_outer_classname\s*=\s*"\(.*\)";/\1/').java))

ALL_TARGETS := $(GO_TARGETS) $(CPP_TARGETS) $(PYTHON_TARGETS) $(JAVA_TARGETS)

BUILD_IMAGE := proio-gen.img
COMMAND_PREFIX := singularity exec -C -H $(PWD) $(BUILD_IMAGE)
GO_TMP_BASE := tmp/go/src
GO_TMP_DIR := $(GO_TMP_BASE)/github.com/decibelcooper

.PHONY: all clean

all: clean $(ALL_TARGETS)
	rm -rf tmp

clean: 
	rm -f $(ALL_TARGETS)
	rm -f $(CPP_HEADERS)
	rm -f $(patsubst %,%__init__.py,$(sort $(dir $(PYTHON_TARGETS))))
	rm -f go-proio/model_imports.go

$(BUILD_IMAGE):
	SINGULARITY_CACHEDIR=/tmp/singularity-cache singularity build $@ docker://dbcooper/proio-gen

.SECONDEXPANSION:
go-proio/%.pb.go: $(SOURCE) $(BUILD_IMAGE)
	$(COMMAND_PREFIX) bash -c "if [ ! -d $(GO_TMP_DIR)/proio ]; then mkdir -p $(GO_TMP_DIR); ln -s $(PWD) $(GO_TMP_DIR); fi"
	$(COMMAND_PREFIX) protoc --gofast_out=$(GO_TMP_BASE) $(patsubst go-proio/%,%,$(basename $(basename $@))).proto
	$(COMMAND_PREFIX) bash -c ". go-proio/addModelImport.sh go-proio/model_imports.go $(patsubst go-proio/%,%,$(basename $(basename $@))).proto"

cpp-proio/src/proio/%.pb.cc: $(SOURCE) $(BUILD_IMAGE)
	$(COMMAND_PREFIX) protoc --cpp_out=cpp-proio/src/proio $(patsubst cpp-proio/src/proio/%,%,$(basename $(basename $@))).proto
	$(COMMAND_PREFIX) bash -c "cd cpp-proio/src && clang-format -i -style=file $(patsubst cpp-proio/src/%,%,$@)"

py-proio/proio/%.py: $(SOURCE) $(BUILD_IMAGE)
	$(COMMAND_PREFIX) protoc --python_out=py-proio/proio $(patsubst py-proio/proio/%_pb2,%,$(basename $(basename $@))).proto
	$(COMMAND_PREFIX) bash -c "echo \"from .$(basename $(@F)) import *\" >> $(@D)/__init__.py"

java-proio/src/main/java/%.java: $(SOURCE) $(BUILD_IMAGE)
	$(COMMAND_PREFIX) protoc --java_out=java-proio/src/main/java $(shell grep -H \"$(basename $(@F))\" $(shell grep -H \"$(patsubst java-proio/src/main/java/%,%,$(@D))\" $(SOURCE) | sed 's/\(.*\):.*/\1/g') | sed 's/\(.*\):.*/\1/g')

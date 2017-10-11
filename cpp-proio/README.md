# Installing
This is a cmake build system.  Protobuf 3.4+ is a required dependency, and ROOT
is an optional dependency.  If Protobuf is installed in a standard location,
the code can be compiled and installed as follows:
```shell
mkdir build && cd build
cmake ../
make
sudo make install
```
With protobuf in a non-standard location, specify the include and library paths
at the `cmake` step like so:
```shell
cmake -DProtobuf_INCLUDE_DIR=<include-path> -DProtobuf_LIBRARY=<lib-path> ../
```

## Open Science Grid compile
```shell
mkdir build && cd build
module load gcc/6.2.0
module load cmake/3.8.0
CC=$(which gcc) CXX=$(which g++) cmake \
	-DProtobuf_INCLUDE_DIR=<include-path> \
	-DProtobuf_LIBRARY=<lib-path> \
	-DCMAKE_INSTALL_PREFIX=<install-path> \
	../
make install
```

# Examples
Please see [the main readme](../README.md) as well as the source code for the
tools in the subdirectories.

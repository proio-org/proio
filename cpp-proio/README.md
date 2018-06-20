# proio for C++
## API
The API documentation is generated using Doxygen, and can be found
[here](https://decibelcooper.github.io/cpp-proio-docs/).

## Installation
### Requirements
* Protobuf 3.1+
* LZ4 1.8+

### Building the code
Standard CMake practices apply, so make sure you have CMake installed.  Create
a build directory (e.g. cpp-proio/build), and `cd` into it.  Then, run `cmake` on
the directory with `CMakeLists.txt` (this directory).
```shell
mkdir cpp-proio/build
cd cpp-proio/build
cmake ../
make
make test
sudo make install
```

If you need to point CMake to dependencies in non-standard locations, please
set the
[`CMAKE_PREFIX_PATH`](https://cmake.org/cmake/help/v3.0/variable/CMAKE_PREFIX_PATH.html)
variable.  For example, if you build and install the required Protobuf and LZ4
libraries into subdirectories of `/opt`, your `cmake` command might look like
the following:
```shell
cmake \
    -DCMAKE_PREFIX_PATH="/opt/protobuf;/opt/lz4" \
    -DCMAKE_INSTALL_PREFIX=/opt/proio \
    ../
```

### Installing LZ4 from github
The following installs to the default installation prefix.
```
git clone https://github.com/lz4/lz4.git && cd lz4
make cmake
sudo cmake --build contrib/cmake_unofficial/.  -- install
```


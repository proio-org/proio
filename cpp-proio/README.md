# proio for C++
## API
The API documentation is generated using Doxygen, and can be found
[here](https://decibelcooper.github.io/cpp-proio-docs/).

## Installation
### Requirements
* Protobuf 3.1+
* LZ4 1.8+

### Installing LZ4 from github
The following installs to the default installation prefix.
```
git clone https://github.com/lz4/lz4.git && cd lz4
make cmake
sudo cmake --build contrib/cmake_unofficial/.  -- install
```


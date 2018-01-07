cmake_minimum_required(VERSION 3.0)
project(proio)

set(CMAKE_CXX_STANDARD 11)

if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE RelWithDebInfo CACHE STRING "..." FORCE)
endif()

# Find Protobuf
find_package(Protobuf REQUIRED)
include_directories(${Protobuf_INCLUDE_DIRS})

# Find lz4
find_path(LZ4_INCLUDE_DIR NAMES lz4frame.h)
find_library(LZ4_LIBRARY NAMES lz4)
include_directories(${LZ4_INCLUDE_DIR})
if(NOT LZ4_LIBRARY)
    message(FATAL_ERROR "lz4 was not found!")
endif()

# Generate Protobuf code
set(proioprotos
    ${PROJECT_SOURCE_DIR}/../proto/proio.proto
    )
protobuf_generate_cpp(protosources protoheaders ${proioprotos})
file(GLOB_RECURSE modelprotos ${PROJECT_SOURCE_DIR}/../model/*.proto)
protobuf_generate_cpp(modelsources modelheaders ${modelprotos})
include_directories(${CMAKE_CURRENT_BINARY_DIR})

# Specify library code
set(librarysources
    ${PROJECT_SOURCE_DIR}/src/event.cc
    ${PROJECT_SOURCE_DIR}/src/reader.cc
    ${PROJECT_SOURCE_DIR}/src/writer.cc
    )
set(libraryheaders
    ${PROJECT_SOURCE_DIR}/src/event.h
    ${PROJECT_SOURCE_DIR}/src/reader.h
    ${PROJECT_SOURCE_DIR}/src/writer.h
    )
include_directories(${PROJECT_SOURCE_DIR}/src)

# Add library target
add_library(proio SHARED
    ${protosources} ${protoheaders}
    ${modelsources} ${modelheaders}
    ${librarysources} ${libraryheaders}
    )
target_link_libraries(proio PUBLIC protobuf lz4)

# Install library and headers
install(TARGETS proio DESTINATION lib)
foreach(header ${protoheaders})
    install(FILES ${header} DESTINATION include/proio/proto/)
endforeach()
foreach(header ${modelheaders})
    install(FILES ${header} DESTINATION include/proio/model/)
endforeach()

# Code tests
enable_testing()

add_executable(pushgetinspect src/tests/pushgetinspect/main.cc)
target_link_libraries(pushgetinspect PUBLIC proio protobuf)
add_test(PushGetInspect pushgetinspect)

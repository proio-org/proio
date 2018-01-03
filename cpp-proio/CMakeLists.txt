cmake_minimum_required(VERSION 3.0)
project(proio)

set(CMAKE_CXX_STANDARD 14)

find_package(Protobuf REQUIRED)
include_directories(${Protobuf_INCLUDE_DIRS})
link_directories(${Protobuf_LIBRARY})

set(proioprotos
    ${PROJECT_SOURCE_DIR}/../proto/proio.proto
    )
protobuf_generate_cpp(protosources protoheaders ${proioprotos})
file(GLOB_RECURSE modelprotos ${PROJECT_SOURCE_DIR}/../model/*.proto)
protobuf_generate_cpp(modelsources modelheaders ${modelprotos})
include_directories(${CMAKE_CURRENT_BINARY_DIR})

add_library(proio SHARED ${protosources} ${protoheaders} ${modelsources} ${modelheaders})
target_link_libraries(proio PUBLIC protobuf)
install(TARGETS proio DESTINATION lib)
foreach(header ${protoheaders})
    install(FILES ${header} DESTINATION include/proto/)
endforeach()
foreach(header ${modelheaders})
    install(FILES ${header} DESTINATION include/model/)
endforeach()

enable_testing()

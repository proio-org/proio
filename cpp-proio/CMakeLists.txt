cmake_minimum_required(VERSION 2.8)
project(proio)

set(libproiofiles
	src/proio/proio.pb.cc
	src/proio/event.cc
	src/proio/reader.cc
	src/proio/writer.cc
)

set(libproioheaders
	src/proio/proio.pb.h
	src/proio/event.h
	src/proio/reader.h
	src/proio/writer.h
)
include_directories(src)

find_package(Protobuf REQUIRED)
include_directories(${Protobuf_INCLUDE_DIRS})

add_library(proio SHARED ${libproiofiles} ${libproioheaders})
target_link_libraries(proio PUBLIC protobuf)
install(TARGETS proio DESTINATION lib)
install(FILES ${libproioheaders} DESTINATION include/proio)

add_executable(proio-ls src/tools/proio-ls/main.cc)
target_link_libraries(proio-ls PUBLIC proio protobuf)
install(TARGETS proio-ls DESTINATION bin)

add_executable(proio-summary src/tools/proio-summary/main.cc)
target_link_libraries(proio-summary PUBLIC proio protobuf)
install(TARGETS proio-summary DESTINATION bin)

add_executable(proio-reserialize src/tools/proio-reserialize/main.cc)
target_link_libraries(proio-reserialize PUBLIC proio protobuf)
install(TARGETS proio-reserialize DESTINATION bin)

find_package(ROOT COMPONENTS RIO Tree)

if(ROOT_FOUND)
    add_executable(proio2root src/tools/proio2root/main.cc)
    target_link_libraries(proio2root PUBLIC proio ${Protobuf_LIBRARIES} ${ROOT_LIBRARIES})
    target_include_directories(proio2root PUBLIC ${Protobuf_INCLUDE_DIRS} ${ROOT_INCLUDE_DIRS})
    install(TARGETS proio2root DESTINATION bin)
endif(ROOT_FOUND)

enable_testing()

add_executable(refderef src/tests/refderef/main.cc)
target_link_libraries(refderef PUBLIC proio protobuf)
add_test(RefDeref refderef)

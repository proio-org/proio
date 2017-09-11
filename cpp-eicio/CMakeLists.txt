cmake_minimum_required(VERSION 2.8)
project(EICIO)

set(libeiciofiles
	src/eicio/eicio.pb.cc
	src/eicio/event.cc
	src/eicio/reader.cc
)

set(libeicioheaders
	src/eicio/eicio.pb.h
	src/eicio/event.h
	src/eicio/reader.h
)
include_directories(src)

find_package(Protobuf)
include_directories(${Protobuf_INCLUDE_DIRS})

add_library(eicio SHARED ${libeiciofiles} ${libeicioheaders})
target_link_libraries(eicio PUBLIC protobuf)
install(TARGETS eicio DESTINATION lib)
install(FILES ${libeicioheaders} DESTINATION include/eicio)

add_executable(eicio-ls src/tools/eicio-ls/main.cc)
target_link_libraries(eicio-ls PUBLIC eicio protobuf)
install(TARGETS eicio-ls DESTINATION bin)

add_executable(eicio-summary src/tools/eicio-summary/main.cc)
target_link_libraries(eicio-summary PUBLIC eicio protobuf)
install(TARGETS eicio-summary DESTINATION bin)

# TODO: make the following optional
find_package(ROOT REQUIRED COMPONENTS RIO Tree)

add_executable(eicio2root src/tools/eicio2root/main.cc)
target_link_libraries(eicio2root PUBLIC eicio ${Protobuf_LIBRARIES} ${ROOT_LIBRARIES})
target_include_directories(eicio2root PUBLIC ${Protobuf_INCLUDE_DIRS} ${ROOT_INCLUDE_DIRS})
install(TARGETS eicio2root DESTINATION bin)

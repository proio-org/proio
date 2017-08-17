cmake_minimum_required(VERSION 2.8)
project(EICIO)

find_package(Protobuf)

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

add_library(eicio SHARED ${libeiciofiles} ${libeicioheaders})
target_link_libraries(eicio PUBLIC protobuf)
install(TARGETS eicio DESTINATION lib)
install(FILES ${libeicioheaders} DESTINATION include/eicio)

add_executable(eicio2root src/tools/eicio2root/main.cc)
target_link_libraries(eicio2root PUBLIC eicio protobuf)
install(TARGETS eicio2root DESTINATION bin)

add_executable(eicio-ls src/tools/eicio-ls/main.cc)
target_link_libraries(eicio-ls PUBLIC eicio protobuf)
install(TARGETS eicio-ls DESTINATION bin)

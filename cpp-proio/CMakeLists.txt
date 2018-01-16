cmake_minimum_required(VERSION 3.0)
project(proio VERSION 0.1.1)

set(CMAKE_CXX_STANDARD 11)

if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE RelWithDebInfo CACHE STRING "..." FORCE)
endif()

# Find Protobuf
find_package(Protobuf 3 REQUIRED)
include_directories(${Protobuf_INCLUDE_DIRS})
get_filename_component(Protobuf_LIBRARY_DIR ${Protobuf_LIBRARY} DIRECTORY)
link_directories(${Protobuf_LIBRARY_DIR})

# Find lz4
find_path(LZ4_INCLUDE_DIR NAMES lz4frame.h)
if(NOT LZ4_INCLUDE_DIR)
    message(FATAL_ERROR "lz4frame.h not found!")
endif()
include_directories(${LZ4_INCLUDE_DIR})
find_library(LZ4_LIBRARY NAMES lz4)
if(NOT LZ4_LIBRARY)
    message(FATAL_ERROR "lz4 was not found!")
endif()
get_filename_component(LZ4_LIBRARY_DIR ${LZ4_LIBRARY} DIRECTORY)
link_directories(${LZ4_LIBRARY_DIR})

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
set(proio_INCLUDE_DIR include)
install(TARGETS proio
    EXPORT proioTargets
    DESTINATION lib
    )
foreach(header ${protoheaders})
    install(FILES ${header}
        DESTINATION ${proio_INCLUDE_DIR}/proio/
        )
endforeach()
foreach(header ${modelheaders})
    install(FILES ${header}
        DESTINATION ${proio_INCLUDE_DIR}/proio/model/
        )
endforeach()
foreach(header ${libraryheaders})
    install(FILES ${header}
        DESTINATION ${proio_INCLUDE_DIR}/proio/
        )
endforeach()

# Install and export targets
install(EXPORT proioTargets
    FILE proioTargets.cmake
    DESTINATION lib/proio
    )

include(CMakePackageConfigHelpers)

set(TARGETS_INSTALL_PATH lib/proio/proioTargets.cmake)
set(proio_LIBRARY lib/libproio.so)
CONFIGURE_PACKAGE_CONFIG_FILE(
    cmake/proioConfig.cmake.in
    ${CMAKE_CURRENT_BINARY_DIR}/proioConfig.cmake
    INSTALL_DESTINATION lib/proio
    PATH_VARS TARGETS_INSTALL_PATH proio_INCLUDE_DIR proio_LIBRARY
    )

write_basic_package_version_file("proioConfigVersion.cmake"
    VERSION ${proio_VERSION}
    COMPATIBILITY SameMajorVersion
    )

install(FILES
    ${CMAKE_CURRENT_BINARY_DIR}/proioConfig.cmake
    ${CMAKE_CURRENT_BINARY_DIR}/proioConfigVersion.cmake
    DESTINATION lib/proio
    )

# Code tests
enable_testing()

add_executable(pushgetinspect src/tests/pushgetinspect/main.cc)
target_link_libraries(pushgetinspect PUBLIC proio protobuf)
add_test(PushGetInspect pushgetinspect)

add_executable(refderef src/tests/refderef/main.cc)
target_link_libraries(refderef PUBLIC proio protobuf)
add_test(RefDeref refderef)

add_executable(tagmanip src/tests/tagmanip/main.cc)
target_link_libraries(tagmanip PUBLIC proio protobuf)
add_test(TagManip tagmanip)

add_executable(scan src/tests/scan/main.cc)
target_link_libraries(scan PUBLIC proio protobuf)
add_test(Scan scan)

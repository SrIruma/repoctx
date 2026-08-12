cmake_minimum_required(VERSION 3.20)
project(repoctx-demo CXX)

find_package(Boost REQUIRED)
find_package(OpenSSL 3.0 REQUIRED)

add_custom_target(build
  COMMAND cmake --build build
  WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
)

add_custom_target(test COMMAND ctest)

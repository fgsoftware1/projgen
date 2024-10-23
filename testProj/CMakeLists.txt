
cmake_minimum_required(VERSION 3.30.5)

# Set the project name and language
project(testProj LANGUAGES CXX)

# Set the project language standard

set(CMAKE_CXX_STANDARD 20)


# Add the executable or library

add_executable(testProj src/main.cpp)


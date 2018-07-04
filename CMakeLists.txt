cmake_minimum_required(VERSION 3.0)

project(GoAppImage)

add_subdirectory(AppImageKit)

set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${PROJECT_BINARY_DIR}/lib)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${PROJECT_BINARY_DIR}/lib)
set(CMAKE_BINARY_OUTPUT_DIRECTORY ${PROJECT_BINARY_DIR}/bin)

include_directories(${C_INCLUDE_PATH})

if(DEFINED ENV{VERBOSE})
  set(GOFLAGS "-v" "-x" "--ldflags" "'-extldflags" "\"-static\"'")
else()
  set(GOFLAGS "--ldflags" "'-extldflags" "\"-static\"'")
endif()

set(GOCMD go build ${GOFLAGS} ${CMAKE_PROJECT_NAME})

target_link_libraries(GoAppImage PRIVATE libappimage_static inotify-tools xdg-basedir dl)

add_custom_target( goinstall ALL
  ${GOCMD}
  COMMENT "Running ${GOCMDPP}"
)
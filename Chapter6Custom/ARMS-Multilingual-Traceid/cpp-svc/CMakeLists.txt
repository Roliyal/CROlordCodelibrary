cmake_minimum_required(VERSION 3.20)
project(cpp_svc CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(httplib CONFIG REQUIRED)
find_package(nlohmann_json CONFIG REQUIRED)
find_package(opentelemetry-cpp CONFIG REQUIRED)

add_executable(cpp-svc main.cpp)

target_link_libraries(cpp-svc PRIVATE
  httplib::httplib
  nlohmann_json::nlohmann_json
  opentelemetry-cpp::api
  opentelemetry-cpp::sdk
  opentelemetry-cpp::trace
)

set(_otel_targets
  opentelemetry-cpp::otlp_http_exporter
  opentelemetry-cpp::otlp_http_client
  opentelemetry-cpp::otlp_recordable
  opentelemetry-cpp::exporter_otlp_http
  opentelemetry-cpp::exporter_otlp_http_client
)

foreach(t IN LISTS _otel_targets)
  if(TARGET ${t})
    message(STATUS "Linking ${t}")
    target_link_libraries(cpp-svc PRIVATE ${t})
  endif()
endforeach()

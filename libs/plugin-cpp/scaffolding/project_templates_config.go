package scaffolding

// CMakeMinimum is the CMake new projects require (PROJECT_IS_TOP_LEVEL).
const CMakeMinimum = "3.25"

var listsTemplate = map[string]string{
	"[[.Dir]]/CMakeLists.txt": `cmake_minimum_required(VERSION ` + CMakeMinimum + `)

# vcpkg: with a vcpkg.json beside this file, libraries come from VCPKG_ROOT.
if(EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/vcpkg.json" AND DEFINED ENV{VCPKG_ROOT} AND NOT DEFINED CMAKE_TOOLCHAIN_FILE)
  set(CMAKE_TOOLCHAIN_FILE "$ENV{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake")
endif()

project([[.Target]] VERSION 0.0.0 LANGUAGES CXX)
set(SMITH_KIND "[[.Kind]]")
[[- if .Stack]]
set(SMITH_STACK "[[.Stack]]")
[[- end]]

# The code, one folder per outcome under src/ (its API under include/).
add_library([[.Library]]
  src/greeting/greet_use_case.cpp
)
target_include_directories([[.Library]] PUBLIC include)
target_compile_features([[.Library]] PUBLIC cxx_std_20)
[[- if eq .Kind "app"]]

add_executable([[.Target]] src/main.cpp)
target_link_libraries([[.Target]] PRIVATE [[.Library]])
[[- end]]

# Tests run when this project is built itself, not when another project
# adds it with add_subdirectory.
if(PROJECT_IS_TOP_LEVEL)
  enable_testing()
  add_executable([[.Target]]_tests tests/greeting/greet_use_case_test.cpp)
  target_link_libraries([[.Target]]_tests PRIVATE [[.Library]])
  add_test(NAME [[.Target]]_tests COMMAND [[.Target]]_tests)
endif()
`,
}

// The greeting slice, a worked example of a vertical slice in C++: a folder
// per outcome, its API under include/<project>/, a file per role, the tests
// mirrored under tests/.
var sliceTemplates = map[string]string{
	"[[.Dir]]/include/[[.Target]]/greeting.hpp": `#pragma once

// The greeting slice's API: a worked example of a slice. Replace it with
// your own outcome.
#include "[[.Target]]/greeting/greet_use_case.hpp"
`,
	"[[.Dir]]/include/[[.Target]]/greeting/greet_use_case.hpp": `#pragma once

#include <string>
#include <string_view>

namespace [[.Target]]::greeting {

// Greets someone by name.
std::string greet(std::string_view name);

}  // namespace [[.Target]]::greeting
`,
	"[[.Dir]]/src/greeting/greet_use_case.cpp": `#include "[[.Target]]/greeting/greet_use_case.hpp"

namespace [[.Target]]::greeting {

std::string greet(std::string_view name) {
  return "Hello, " + std::string(name) + "!";
}

}  // namespace [[.Target]]::greeting
`,
	"[[.Dir]]/tests/greeting/greet_use_case_test.cpp": `#include "[[.Target]]/greeting/greet_use_case.hpp"

#include <cstdlib>
#include <iostream>

int main() {
  if ([[.Target]]::greeting::greet("world") != "Hello, world!") {
    std::cerr << "greet: unexpected greeting\n";
    return EXIT_FAILURE;
  }
  return EXIT_SUCCESS;
}
`,
}

var appTemplates = map[string]string{
	"[[.Dir]]/src/main.cpp": `#include "[[.Target]]/greeting.hpp"

#include <iostream>

int main() {
  std::cout << [[.Target]]::greeting::greet("world") << '\n';
  return 0;
}
`,
}

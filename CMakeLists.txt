# CMakeLists.txt for selfsign-path-tool release automation
# Windows-compatible build system

cmake_minimum_required(VERSION 3.15)
project(selfsign-path-tool)

# Detect version from git if not specified
if(VERSION STREQUAL "unknown" OR NOT DEFINED VERSION)
	execute_process(
		COMMAND git tag --sort=-version:refname
		OUTPUT_VARIABLE GIT_TAGS
		OUTPUT_STRIP_TRAILING_WHITESPACE
	)
	
	string(REPLACE "\n" ";" GIT_TAGS_LIST "${GIT_TAGS}")
	
	foreach(TAG ${GIT_TAGS_LIST})
		if(TAG MATCHES "^v?[0-9]+\\.[0-9]+\\.[0-9]+")
			string(REGEX REPLACE "^v" "" DETECTED_VERSION "${TAG}")
			set(VERSION "${DETECTED_VERSION}")
			break()
		endif()
	endforeach()
	
	if(VERSION STREQUAL "unknown")
		message(WARNING "No valid semantic version tag found in git, using default version 'unknown'")
	else()
		message(STATUS "Using detected git version: ${VERSION}")
	endif()
endif()

# Default variables (can be overridden with -D when running cmake)
set(VERSION "${VERSION}" CACHE STRING "Version number" FORCE)
set(TAG_NAME "v${VERSION}" CACHE STRING "Git tag name")
set(DIST_DIR "${CMAKE_CURRENT_SOURCE_DIR}/dist" CACHE PATH "Distribution directory")
set(VERSIONED_SCRIPT "${DIST_DIR}/selfsign-path-v${VERSION}.ps1" CACHE STRING "Versioned script name")

# Rest of the file remains unchanged...

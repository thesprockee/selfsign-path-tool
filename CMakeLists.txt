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

# Print configuration information
message(STATUS "=== selfsign-path-tool Build Configuration ===")
message(STATUS "Version: ${VERSION}")
message(STATUS "Tag Name: ${TAG_NAME}")
message(STATUS "Distribution Directory: ${DIST_DIR}")
message(STATUS "Versioned Script: ${VERSIONED_SCRIPT}")

# Ensure PowerShell is available
find_program(POWERSHELL_EXECUTABLE
	NAMES pwsh powershell
	DOC "PowerShell executable"
)

if(NOT POWERSHELL_EXECUTABLE)
	message(FATAL_ERROR "PowerShell not found. Please install PowerShell Core (pwsh) or Windows PowerShell.")
endif()

message(STATUS "PowerShell executable: ${POWERSHELL_EXECUTABLE}")

# Create distribution directory
file(MAKE_DIRECTORY ${DIST_DIR})

# Custom target to create versioned script
add_custom_target(create-versioned-script
	COMMAND ${POWERSHELL_EXECUTABLE} -ExecutionPolicy Bypass -File ${CMAKE_CURRENT_SOURCE_DIR}/scripts/create-versioned-script.ps1 -Version ${VERSION}
	WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
	COMMENT "Creating versioned script for version ${VERSION}"
	VERBATIM
)

# Custom target to generate changelog
add_custom_target(generate-changelog
	COMMAND ${POWERSHELL_EXECUTABLE} -ExecutionPolicy Bypass -File ${CMAKE_CURRENT_SOURCE_DIR}/scripts/generate-changelog.ps1 -Version ${VERSION} -TagName ${TAG_NAME}
	WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
	DEPENDS create-versioned-script
	COMMENT "Generating changelog for version ${VERSION}"
	VERBATIM
)

# Custom target to sign script (optional, based on environment variables)
add_custom_target(sign-script
	COMMAND ${POWERSHELL_EXECUTABLE} -ExecutionPolicy Bypass -File ${CMAKE_CURRENT_SOURCE_DIR}/scripts/sign-script.ps1 -VersionedScript ${CMAKE_CURRENT_SOURCE_DIR}/selfsign-path-v${VERSION}.ps1
	WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
	DEPENDS create-versioned-script
	COMMENT "Signing script (if certificates available)"
	VERBATIM
)

# Main distribution target - creates all release artifacts
add_custom_target(dist
	DEPENDS generate-changelog sign-script
	COMMENT "Creating distribution package for version ${VERSION}"
)

# Clean target to remove generated files
add_custom_target(clean-dist
	COMMAND ${CMAKE_COMMAND} -E remove -f 
		${CMAKE_CURRENT_SOURCE_DIR}/selfsign-path-v${VERSION}.ps1
		${CMAKE_CURRENT_SOURCE_DIR}/RELEASE_NOTES.md
	COMMAND ${CMAKE_COMMAND} -E remove_directory ${DIST_DIR}
	COMMENT "Cleaning generated distribution files"
)

# Help target
add_custom_target(help-dist
	COMMAND ${CMAKE_COMMAND} -E echo "=== selfsign-path-tool Build Targets ==="
	COMMAND ${CMAKE_COMMAND} -E echo "dist              - Create all release artifacts"
	COMMAND ${CMAKE_COMMAND} -E echo "create-versioned-script - Create versioned PowerShell script"
	COMMAND ${CMAKE_COMMAND} -E echo "generate-changelog      - Generate changelog and release notes"
	COMMAND ${CMAKE_COMMAND} -E echo "sign-script            - Sign the script if certificates available"
	COMMAND ${CMAKE_COMMAND} -E echo "clean-dist             - Clean generated files"
	COMMAND ${CMAKE_COMMAND} -E echo "help-dist              - Show this help message"
	COMMAND ${CMAKE_COMMAND} -E echo ""
	COMMAND ${CMAKE_COMMAND} -E echo "Environment Variables:"
	COMMAND ${CMAKE_COMMAND} -E echo "VERSION              - Version number required"
	COMMAND ${CMAKE_COMMAND} -E echo "TAG_NAME             - Git tag name defaults to vVERSION"
	COMMAND ${CMAKE_COMMAND} -E echo "SIGNING_CERT         - Base64 encoded signing certificate optional"
	COMMAND ${CMAKE_COMMAND} -E echo "SIGNING_CERT_PASSWORD - Password for signing certificate optional"
	COMMENT "Showing help information"
)

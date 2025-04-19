# ===================================================================
# Copyright (c) 2024, HIT Authors
# All rights reserved.
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions
# are met:
# 
# 1. Redistributions of source code must retain the above copyright
# notice, this list of conditions and the following disclaimer.
# 2. Redistributions in binary form must reproduce the above copyright
# notice, this list of conditions and the following disclaimer in the
# documentation and/or other materials provided with the
# distribution.
# 3. All advertising materials mentioning features or use of this software
# must display the following acknowledgement:
# This product includes software developed by the xxx Group. and
# its contributors.
# 4. Neither the name of the Group nor the names of its contributors may
# be used to endorse or promote products derived from this software
# without specific prior written permission.
# 
# THIS SOFTWARE IS PROVIDED BY Wanyou Wang,GROUP AND CONTRIBUTORS
# ===================================================================
# Author: Wanyou Wang

DBG_MAKEFILE ?=
ifeq ($(DBG_MAKEFILE),1)
    $(warning ***** starting Makefile for goal(s) "$(MAKECMDGOALS)")
    $(warning ***** $(shell date))
else
    # If we're not debugging the Makefile, don't echo recipes.
    MAKEFLAGS += -s
endif

# Define optional user-input variables so `make --warn-undefined-variables` works.
# Usage is described below with each target that respects these
PRINT_HELP ?=
WHAT ?=
TESTS ?=
BRANCH ?=

# This controls the verbosity of the build.  Higher numbers mean more output.
FRAMEWORK_VERBOSE ?= 3

define ALL_HELP_INFO
# Build code.
#
# Args:
#   WHAT: Directory or Go package names to build.  If any of these directories
#   has a 'main' package, the build will produce executable files under
#   $(OUT_DIR)/bin.  If not specified, "everything" will be built.
#     "vendor/<module>/<path>" is accepted as alias for "<module>/<path>".
#     "ginkgo" is an alias for the ginkgo CLI.
#   GOFLAGS: Extra flags to pass to 'go' when building.
#   GOLDFLAGS: Extra linking flags passed to 'go' when building.
#   GOGCFLAGS: Additional go compile flags passed to 'go' when building.
#   DBG: If set to "1", build with optimizations disabled for easier
#     debugging.  Any other value is ignored.
#
# Example:
#   make
#   make all
#   make all WHAT=cmd/resourcelet GOFLAGS=-v
#   make all DBG=1
#     Note: Specify DBG=1 for building unstripped binaries, which allows you to
#     use code debugging tools like delve. When DBG is unspecified, it defaults
#     to "-s -w" which strips debug information.
endef
.PHONY: all
ifeq ($(PRINT_HELP),y)
all:
	echo "$$ALL_HELP_INFO"
else
all:
	build/make-rules/build.sh $(WHAT)
endif
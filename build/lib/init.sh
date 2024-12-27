#!/usr/bin/env bash

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

set -o errexit
set -o nounset
set -o pipefail

# Short-circuit if init.sh has already been sourced
[[ $(type -t framework::init::loaded) == function ]] && return 0

# Unset CDPATH so that path interpolation can work correctly
unset CDPATH

# FIXME(dims): Note that here we assume that if GOFLAGS are already set we
# leave them as-is and not try to add providerless to it. So if you
# really need to set your own GOFLAGS, ensure you add "providerless" explicitly
if [[ "${FRAMEWORK_PROVIDERLESS:-"n"}" == "y" ]]; then
  export GOFLAGS=${GOFLAGS:-"-tags=providerless"}
fi

# The root of the build/dist directory
FRAMEWORK_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

# Where output goes.  We should avoid redefining these anywhere else.
#
# FRAMEWORK_OUTPUT: the root directory (absolute) where this build should drop any
#     files (subdirs are encouraged).
# FRAMEWORK_OUTPUT_BIN: the directory in which compiled binaries will be placed,
#     under OS/ARCH specific subdirs
# THIS_PLATFORM_BIN: a symlink to the output directory for binaries built for
#     the current host platform (e.g. build/test tools).
#
# Compat: The FRAMEWORK_OUTPUT_SUBPATH variable is sometimes passed in by callers.
# If it is specified, we'll use it in FRAMEWORK_OUTPUT.
_FRAMEWORK_OUTPUT_SUBPATH="${FRAMEWORK_OUTPUT_SUBPATH:-_output/local}"
export FRAMEWORK_OUTPUT="${FRAMEWORK_ROOT}/${_FRAMEWORK_OUTPUT_SUBPATH}"
export FRAMEWORK_OUTPUT_BIN="${FRAMEWORK_OUTPUT}/bin"
export THIS_PLATFORM_BIN="${FRAMEWORK_ROOT}/_output/bin"

# This controls rsync compression. Set to a value > 0 to enable rsync
# compression for build container
FRAMEWORK_RSYNC_COMPRESS="${FRAMEWORK_RSYNC_COMPRESS:-0}"

# Set no_proxy for localhost if behind a proxy, otherwise,
# the connections to localhost in scripts will time out
export no_proxy="127.0.0.1,localhost${no_proxy:+,${no_proxy}}"

# source "${FRAMEWORK_ROOT}/build/lib/util.sh"
source "${FRAMEWORK_ROOT}/build/lib/logging.sh"

framework::log::install_errexit

source "${FRAMEWORK_ROOT}/build/lib/golang.sh"

# list of all available group versions.  This should be used when generated code
# or when starting an API server that you want to have everything.
# most preferred version for a group should appear first
FRAMEWORK_AVAILABLE_GROUP_VERSIONS="${FRAMEWORK_AVAILABLE_GROUP_VERSIONS:-\
}"

# not all group versions are exposed by the server.  This list contains those
# which are not available so we don't generate clients or swagger for them
FRAMEWORK_NONSERVER_GROUP_VERSIONS="
"
export FRAMEWORK_NONSERVER_GROUP_VERSIONS

function framework::readlinkdashf {
  # run in a subshell for simpler 'cd'
  (
    if [[ -d "${1}" ]]; then # This also catch symlinks to dirs.
      cd "${1}"
      pwd -P
    else
      cd "$(dirname "${1}")"
      local f
      f=$(basename "${1}")
      if [[ -L "${f}" ]]; then
        readlink "${f}"
      else
        echo "$(pwd -P)/${f}"
      fi
    fi
  )
}

framework::realpath() {
  if [[ ! -e "${1}" ]]; then
    echo "${1}: No such file or directory" >&2
    return 1
  fi
  framework::readlinkdashf "${1}"
}

# Marker function to indicate init.sh has been fully sourced
framework::init::loaded() {
  return 0
}


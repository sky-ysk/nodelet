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

readonly FRAMEWORK_GOPATH="${FRAMEWORK_GOPATH:-"${FRAMEWORK_OUTPUT}/go"}"
export FRAMEWORK_GOPATH

# Asks golang what it thinks the host platform is. The go tool chain does some
# slightly different things when the target platform matches the host platform.
framework::golang::host_platform() {
  echo "$(go env GOHOSTOS)/$(go env GOHOSTARCH)"
}

framework::golang::setup_env() {
  export GOPATH="${FRAMEWORK_GOPATH}"

  # If these are not set, set them now.  This ensures that any subsequent
  # scripts we run (which may call this function again) use the same values.
  export GOCACHE="${GOCACHE:-"${FRAMEWORK_GOPATH}/cache/build"}"
  export GOMODCACHE="${GOMODCACHE:-"${FRAMEWORK_GOPATH}/cache/mod"}"

  # Make sure our own Go binaries are in PATH.
  export PATH="${FRAMEWORK_GOPATH}/bin:${PATH}"

  # Unset GOBIN in case it already exists in the current session.
  # Cross-compiles will not work with it set.
  unset GOBIN

  # Turn on modules and workspaces (both are default-on).
  unset GO111MODULE
  unset GOWORK
}

framework::golang::build_binaries() {
    V=2 framework::log::info "Go version: $(GOFLAGS='' go version)"

    local -a build_args
    build_args=(
    )

    V=2 framework::log::info "Start to build binary"
    go install "${build_args[@]}" "$@"
}

framework::golang::place_bins() {
    V=2 framework::log::status "Placing binaries"
    local platform
    platform="linux/amd64"
    local platform_src="/${platform//\//_}"

    rm -f "${THIS_PLATFORM_BIN}"
    mkdir -p "$(dirname "${THIS_PLATFORM_BIN}")"
    # ln -s "${FRAMEWORK_OUTPUT_BIN}/${platform}" "${THIS_PLATFORM_BIN}"

    # V=3 framework::log::status "Placing binaries for ${platform} in ${FRAMEWORK_OUTPUT_BIN}/${platform}"
    # local full_binpath_src="${FRAMEWORK_GOPATH}/bin${platform_src}"
    # if [[ -d "${full_binpath_src}" ]]; then
    #   mkdir -p "${FRAMEWORK_OUTPUT_BIN}/${platform}"
    #   find "${full_binpath_src}" -maxdepth 1 -type f -exec \
    #     rsync -pc {} "${FRAMEWORK_OUTPUT_BIN}/${platform}" \;
    # fi
}


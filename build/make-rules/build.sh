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

FRAMEWORK_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..
FRAMEWORK_VERBOSE="${FRAMEWORK_VERBOSE:-1}"
source "${FRAMEWORK_ROOT}/build/lib/init.sh"

framework::golang::setup_env
framework::golang::build_binaries "$@"
framework::golang::place_bins

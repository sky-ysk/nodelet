#!/bin/bash


make all WHAT=./cmd/proxy FRAMEWORK_VERBOSE=3
make all WHAT=./cmd/apiserver FRAMEWORK_VERBOSE=3
make all WHAT=./cmd/nodelet FRAMEWORK_VERBOSE=3
make all WHAT=./cmd/scheduler FRAMEWORK_VERBOSE=3
make all WHAT=./cmd/registry FRAMEWORK_VERBOSE=3
#!/usr/bin/env bash

export PROJECT_ROOT=$(pwd)
export TESTING=true

go test $@

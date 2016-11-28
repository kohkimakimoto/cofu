#!/usr/bin/env bash
set -eu

repo_dir=$(pwd)
go version

export GOPATH="/tmp/dev"

go test $GOTEST_FLAGS $(go list ./... | grep -v vendor)

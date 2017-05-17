#!/usr/bin/env bash
set -e

if [ -z "$IN_CONTAINER" ]; then
  GOTEST_FLAGS=${GOTEST_FLAGS:--cover -timeout=360s}
  DOCKER_IMAGE=${DOCKER_IMAGE:-'kohkimakimoto/golang:centos7'}

  test_dir=$(cd $(dirname $0); pwd)
  cd "$test_dir/.."
  repo_dir=$(pwd)

  echo "Running tests (docker image: $DOCKER_IMAGE) (flags: $GOTEST_FLAGS)..."
  echo "Starting a docker container..."
  exec docker run \
    --env DOCKER_IMAGE="${DOCKER_IMAGE}" \
    --env GOTEST_FLAGS="${GOTEST_FLAGS}" \
    --env IN_CONTAINER=1 \
    -v $repo_dir:/tmp/dev/src/github.com/kohkimakimoto/cofu \
    -w /tmp/dev/src/github.com/kohkimakimoto/cofu \
    --rm \
    ${DOCKER_IMAGE} \
    bash ./test/test.sh
fi

repo_dir=$(pwd)
go version
export GOPATH="/tmp/dev"
go test $GOTEST_FLAGS $(go list ./... | grep -v vendor)


  


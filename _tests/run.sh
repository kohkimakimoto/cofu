#!/usr/bin/env bash
set -e

if [ "${TERM:-dumb}" != "dumb" ]; then
    txtunderline=$(tput sgr 0 1)     # Underline
    txtbold=$(tput bold)             # Bold
    txtred=$(tput setaf 1)           # red
    txtgreen=$(tput setaf 2)         # green
    txtyellow=$(tput setaf 3)        # yellow
    txtblue=$(tput setaf 4)          # blue
    txtreset=$(tput sgr0)            # Reset
else
    txtunderline=""
    txtbold=""
    txtred=""
    txtgreen=""
    txtyellow=""
    txtblue=$""
    txtreset=""
fi

go_version="1.7"

if [ -z "$DOCKER_IMAGE" ]; then
  # no docker
  echo "Running integration tests..."

  case $(uname) in
    Darwin)

      ;;
    *)
      user=`whoami`
      if [ $user != "root" ]; then
          echo "you need to run it on the 'root' user." 1>&2
          exit 1
      fi
      ;;
  esac

  repo_dir=$(cd $(dirname $0)/..; pwd)
  cd $repo_dir

  go_archive="$repo_dir/_tests/cache/go${go_version}.linux-amd64.tar.gz"
  if [ ! -e "$go_archive" ]; then
    echo "downloading go linux archive..."
    cd $(dirname $go_archive)
    curl -O -L https://storage.googleapis.com/golang/go1.7.linux-amd64.tar.gz
    cd -
  fi

  for platform in 'centos:centos5' 'centos:centos6' 'centos:centos7' 'debian:7' 'debian:8' 'ubuntu:12.04' 'ubuntu:14.04' 'ubuntu:16.04'; do
    docker run \
      --privileged  \
      --env DOCKER_IMAGE=${platform}  \
      -v $repo_dir:/tmp/dev/src/github.com/kohkimakimoto/cofu \
      -w /tmp/dev/src/github.com/kohkimakimoto/cofu \
      ${platform} \
      bash ./_tests/run.sh &&:

    if [ $? -eq 0 ]; then
      echo "${txtgreen}$platform OK${txtreset}"
    else
      echo "${txtred}$platform FAIL${txtreset}"
      exit 1
    fi
  done

  echo "Removing tarminated containers..."
  docker rm `docker ps -a -q`
  echo "${txtgreen}Finished successfully all integration tests!${txtreset}"

else
  echo "Running tests on '$DOCKER_IMAGE'..."

  # in a docker container.
  repo_dir=$(cd $(dirname $0)/..; pwd)
  cd $repo_dir

  # https://golang.org/doc/install
  cd _tests/cache
  tar -C /usr/local -xzf go1.7.linux-amd64.tar.gz
  export PATH=$PATH:/usr/local/go/bin
  export GOPATH=/tmp/dev
  export RUN_INTEGRATION_TEST=1
  cd -

  go version

  # now runs only integration test.
  go test -v .
  # go test -v $(go list ./... | grep -v vendor)
fi

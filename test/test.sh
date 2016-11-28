#!/usr/bin/env bash
set -eu

indent() {
    local n="${1:-4}"
    local p=""
    for i in `seq 1 $n`; do
        p="$p "
    done;

    local c="s/^/$p/"
    case $(uname) in
      Darwin) sed -l "$c";; # mac/bsd sed: -l buffers on line boundaries
      *)      sed -u "$c";; # unix/gnu sed: -u unbuffered (arbitrary) chunks of data
    esac
}

prefix() {
  local p="${1:-prefix}"
  local c="s/^/$p/"
  case $(uname) in
    Darwin) sed -l "$c";; # mac/bsd sed: -l buffers on line boundaries
    *)      sed -u "$c";; # unix/gnu sed: -u unbuffered (arbitrary) chunks of data
  esac
}

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

GOTEST_FLAGS=${GOTEST_FLAGS:--cover -timeout=360s}

test_dir=$(cd $(dirname $0); pwd)
cd "$test_dir/.."
repo_dir=$(pwd)

echo "--> Running tests (flags: $GOTEST_FLAGS)..."

for DOCKER_IMAGE in 'kohkimakimoto/golang:centos7' 'kohkimakimoto/golang:centos6'; do
  container_name_part=$(echo $DOCKER_IMAGE | perl -pe "s/[:\/]/-/g;")
  DOCKER_CONTAINER_NAME="$container_name_part-$(date +%s)"

  trap "echo '--> Terminating a container...' && \
    docker rm $DOCKER_CONTAINER_NAME 2>&1 | prefix '    Deleted: ' && \
    echo '--> Done.'" EXIT HUP INT QUIT TERM

  echo "--> Starting a docker container '$DOCKER_CONTAINER_NAME'.."
  echo "    Docker Image: ${txtbold}${txtyellow}$DOCKER_IMAGE${txtreset}"
  docker run \
    --name $DOCKER_CONTAINER_NAME \
    --env DOCKER_IMAGE="${DOCKER_IMAGE}" \
    --env GOTEST_FLAGS="${GOTEST_FLAGS}" \
    -v $repo_dir:/tmp/dev/src/github.com/kohkimakimoto/cofu \
    -w /tmp/dev/src/github.com/kohkimakimoto/cofu \
    ${DOCKER_IMAGE} \
    bash ./test/run_test.sh 2>&1 | indent; status=${PIPESTATUS[0]}

  if [ $status -eq 0 ]; then
    echo "    ${txtgreen}$DOCKER_IMAGE OK${txtreset}"
  else
    echo "    ${txtred}$DOCKER_IMAGE FAIL${txtreset}"
    exit 1
  fi 
  
done


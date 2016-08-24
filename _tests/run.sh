#!/usr/bin/env bash
set -e

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

  # Create dist binaries for testing.
  make dist
  unzip -d ./_build/dist ./_build/dist/cofu_linux_amd64.zip
  trap "rm -f ./_build/dist/cofu" 0

  for platform in 'centos:centos6' 'centos:centos7' 'debian:7' 'debian:8'; do
    docker run \
      --privileged  \
      --env DOCKER_IMAGE=${platform}  \
      -v $repo_dir:/tmp/cofu \
      -w /tmp/cofu \
      ${platform} \
      bash ./_tests/run.sh &&:

    if [ $? -eq 0 ]; then
      echo "${txtgreen}OK${txtreset}"
    else
      echo "${txtred}FAIL${txtreset}"
      exit 1
    fi
  done

  echo "Removing tarminated containers..."
  docker rm `docker ps -a -q`
  echo "Finished integration tests!"

else
  # in a docker container.
  echo "Running tests on '$DOCKER_IMAGE'..."

  repo_dir=$(cd $(dirname $0)/..; pwd)
  cd $repo_dir

  # copy cofu binary to the /usr/local/bin/cofu
  cp -p ${repo_dir}/_build/dist/cofu /usr/local/bin/cofu

  # show cofu version.
  cofu -v

  cd _tests
  cofu ./test_resource_directory/recipe.lua
  cofu ./test_resource_execute/recipe.lua

fi




# prj_dir=$(cd $(dirname $0)/..; pwd)
# cd $prj_dir
#
# user=`whoami`
# if [ $user != "root" ]; then
#     echo "you need to run it on the 'root' user." 1>&2
#     exit 1
# fi
#
# . /etc/profile
# . .envrc
#
# export COFU_INTEGRATION_TEST=1
#
# go test . -v -cover

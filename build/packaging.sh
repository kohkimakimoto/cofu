#!/usr/bin/env bash
set -eu

# Get the directory path.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
build_dir="$( cd -P "$( dirname "$SOURCE" )/" && pwd )"
repo_dir="$(cd $build_dir/.. && pwd)"

# Move the parent (repository) directory
cd "$repo_dir"

# Check if it has loaded .envrc by direnv.
if [ -z ${DIRENV_DIR+x} ]; then
    if [ -e "./.envrc" ]; then
        source ./.envrc
    fi
fi

source $build_dir/config

echo "Building RPM packages..."
cd $build_dir/packaging/rpm
for image in 'kohkimakimoto/rpmbuild:el5' 'kohkimakimoto/rpmbuild:el6' 'kohkimakimoto/rpmbuild:el7'; do
    docker run \
        --env DOCKER_IMAGE=${image}  \
        --env PRODUCT_NAME=${PRODUCT_NAME}  \
        --env PRODUCT_VERSION=${PRODUCT_VERSION}  \
        --env COMMIT_HASH=${COMMIT_HASH}  \
        -v $repo_dir:/tmp/repo \
        -w /tmp/repo \
        --rm \
        ${image} \
        bash ./build/packaging/rpm/run.sh
done

cd "$repo_dir"

echo "Results:"
ls -hl "$build_dir/dist"

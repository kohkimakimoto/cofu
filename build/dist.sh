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

echo "Removing old files."
rm -rf $build_dir/dist/*

COMMIT_HASH=`git log --pretty=format:%H -n 1`

echo "Building dev binary..."
echo "PRODUCT_NAME: $PRODUCT_NAME"
echo "PRODUCT_VERSION: $PRODUCT_VERSION"
echo "COMMIT_HASH: $COMMIT_HASH"

gox \
    -os="linux darwin" \
    -arch="amd64" \
    -ldflags=" -w \
        -X github.com/kohkimakimoto/$PRODUCT_NAME/$PRODUCT_NAME.CommitHash=$COMMIT_HASH \
        -X github.com/kohkimakimoto/$PRODUCT_NAME/$PRODUCT_NAME.Version=$PRODUCT_VERSION" \
    -output "$build_dir/dist/${PRODUCT_NAME}_{{.OS}}_{{.Arch}}" \
    ./cmd/${PRODUCT_NAME}

echo "Packaging to zip archives..."

cd "$build_dir/dist"
echo "Packaging darwin binaries" | indent
mv ${PRODUCT_NAME}_darwin_amd64 ${PRODUCT_NAME} && zip ${PRODUCT_NAME}_darwin_amd64.zip ${PRODUCT_NAME}  | indent && rm ${PRODUCT_NAME}
echo "Packaging linux binaries" | indent
mv ${PRODUCT_NAME}_linux_amd64 ${PRODUCT_NAME} && zip ${PRODUCT_NAME}_linux_amd64.zip ${PRODUCT_NAME}  | indent && rm ${PRODUCT_NAME}
#echo "Packaging windows binaries" | indent
#mv ${PRODUCT_NAME}_windows_amd64.exe ${PRODUCT_NAME}.exe && zip ${PRODUCT_NAME}_windows_amd64.zip ${PRODUCT_NAME}.exe | indent && rm ${PRODUCT_NAME}.exe

cd "$repo_dir"

echo "Results:"
ls -hl "$build_dir/dist"


#!/usr/bin/env bash
set -eu

# Get the directory path.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
build_dir="$( cd -P "$( dirname "$SOURCE" )/" && pwd )"
repo_dir="$(cd $build_dir/.. && pwd)"

# Move the parent (repository) directory
cd "$repo_dir"

echo "Cleaning files."
rm -rf $build_dir/dev/*
rm -rf $build_dir/dist/*

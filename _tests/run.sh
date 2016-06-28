#!/usr/bin/env bash

prj_dir=$(cd $(dirname $0)/..; pwd)
cd $prj_dir

user=`whoami`
if [ $user != "root" ]; then
    echo "you need to run it on the 'root' user." 1>&2
    exit 1
fi

. /etc/profile
. .envrc

export COFU_INTEGRATION_TEST=1

go test . -v -cover

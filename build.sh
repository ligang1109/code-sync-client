#!/bin/bash

if [ $# -eq 1 ]
then
    if [ ! $1 = "linux" ]
    then
        echo "Usage: $0 [linux]"
    fi

    export CGO_ENABLED=0
    export GOOS=linux
    export GOARCH=amd64
fi

curDir=`dirname $0`
cd $curDir
prjHome=`pwd`

if [ ! -d $prjHome/bin ]
then
    mkdir $prjHome/bin
fi

cd src

go build -o code-sync-client main/client/main.go
mv code-sync-client $prjHome/bin

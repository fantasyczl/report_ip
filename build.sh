#!/usr/bin/env bash

set -xe

APP=report_ip

if [ "$GOOS" != "" ]; then
    export GOOS=$GOOS
fi

if [ "$GOARCH" != "" ]; then
    export GOARCH=$GOARCH
fi

BIN_NAME="$APP-$GOOS-$GOARCH"

rm -fr ./output
mkdir -p output

go build -o output/${BIN_NAME} main.go

echo "compile successfully"

cp -fR ./conf output/conf
cd output

echo "packaging..."

tar -czvf ${BIN_NAME}.tar.gz ./conf $BIN_NAME

echo "package has beed complete"

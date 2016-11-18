#!/bin/bash
rm -rf src
mkdir -p src/dht
cp *.go src/dht/
cp -r sample src/
export set GOPATH=`pwd`
go build sample/spider/


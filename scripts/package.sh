#!/bin/bash

set -xe

rm -rf apt-cnb
mkdir apt-cnb
mkdir apt-cnb/bin

export GOARCH="amd64"
export GOOS="linux"
go build -o apt-cnb/bin/detect cmd/detect/main.go
go build -o apt-cnb/bin/build cmd/build/main.go

cp -v buildpack.toml apt-cnb/buildpack.toml

#!/bin/bash

## check if not linux
if [ "$(uname -s)" != "Linux" ]; then
  CC=x86_64-unknown-linux-gnu-gcc CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build .
  exit $?
fi

go build .

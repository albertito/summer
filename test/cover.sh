#!/bin/bash

set -e
cd "$(realpath "$(dirname "$0")" )"

export GOCOVERDIR="$PWD/.cover/"
rm -rf "${GOCOVERDIR?}"
mkdir -p "${GOCOVERDIR?}"

export BUILDARGS="-cover -covermode=count"

# Coverage tests require Go >= 1.20.
go version

go test -cover ../... -covermode=count -args -test.gocoverdir="${GOCOVERDIR?}"

./test.sh

go tool covdata percent -i="${GOCOVERDIR?}"
go tool covdata textfmt \
	-i="${GOCOVERDIR?}" -o="${GOCOVERDIR?}/cover.txt"
go tool cover \
	-html="${GOCOVERDIR?}/cover.txt" -o="${GOCOVERDIR?}/summer.html"

echo "file://${GOCOVERDIR?}/summer.html"


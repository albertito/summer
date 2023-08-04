#!/bin/bash

set -e
cd $(realpath "$(dirname "$0")" )

# shellcheck disable=SC2086
( cd ..; go build $BUILDARGS -o summer . )

TARGETS="${@:-./*.t}"

cram3 $TARGETS

#!/bin/bash

set -e
cd "$(realpath "$(dirname "$0")" )"

# shellcheck disable=SC2086
( cd ..; go build $BUILDARGS -o summer . )

TARGETS="${*:-./*.t}"

# shellcheck disable=SC2086
cram3 $TARGETS

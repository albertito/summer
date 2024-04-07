#!/bin/bash

set -e
cd "$(realpath "$(dirname "$0")" )"

export BUILDARGS="-race"
exec ./test.sh

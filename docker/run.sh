#!/bin/bash
set -eux

ROOT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/.. && pwd )"
cd $ROOT_DIR

PROJECT_VERSION=$(./godelw project-version)

docker run -it -v `pwd`/config:/secrets -p 8000:8000 computetools/sc-metrics:latest

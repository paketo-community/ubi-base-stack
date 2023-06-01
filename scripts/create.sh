#!/usr/bin/env bash

set -eu
set -o pipefail

readonly PROG_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOT_DIR="$(cd "${PROG_DIR}/.." && pwd)"

bash $PROG_DIR/create.sh
STACK_DIR=${ROOT_DIR}/stack-nodejs-16 BUILD_DIR=${ROOT_DIR}/build-nodejs-16 bash $PROG_DIR/create.sh
STACK_DIR=${ROOT_DIR}/stack-nodejs-18 BUILD_DIR=${ROOT_DIR}/build-nodejs-18 bash $PROG_DIR/create.sh


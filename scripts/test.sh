#!/usr/bin/env bash

set -eu
set -o pipefail

readonly PROG_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly STACK_DIR="$(cd "${PROG_DIR}/.." && pwd)"
readonly OUTPUT_DIR="${STACK_DIR}/build"
readonly OUTPUT_DIR_NODEJS16="${STACK_DIR}/build-nodejs-16"
readonly OUTPUT_DIR_NODEJS18="${STACK_DIR}/build-nodejs-18"
readonly OUTPUT_DIR_NODEJS20="${STACK_DIR}/build-nodejs-20"
readonly OUTPUT_DIR_JAVA8="${STACK_DIR}/build-java-8"
readonly OUTPUT_DIR_JAVA11="${STACK_DIR}/build-java-11"
readonly OUTPUT_DIR_JAVA17="${STACK_DIR}/build-java-17"
readonly OUTPUT_DIR_JAVA21="${STACK_DIR}/build-java-21"

# shellcheck source=SCRIPTDIR/.util/tools.sh
source "${PROG_DIR}/.util/tools.sh"

# shellcheck source=SCRIPTDIR/.util/print.sh
source "${PROG_DIR}/.util/print.sh"

function main() {
  local clean token
  clean="false"
  token=""

  local regport create_registry
  regport=""
  create_registry="true"

  while [[ "${#}" != 0 ]]; do
    case "${1}" in
      --help|-h)
        shift 1
        usage
        exit 0
        ;;

      --clean|-c)
        shift 1
        clean="true"
        ;;

      --token|-t)
        token="${2}"
        shift 2
        ;;

    --regport | -p)
      regport="${2}"
      shift 2
      ;;

      "")
        # skip if the argument is empty
        shift 1
        ;;

      *)
        util::print::error "unknown argument \"${1}\""
    esac
  done

  tools::install "${token}"

  if [[ "${clean}" == "true" ]]; then
    util::print::title "Cleaning up preexisting stack archives..."
    rm -rf "${OUTPUT_DIR}"
    rm -rf "${OUTPUT_DIR_NODEJS16}"
    rm -rf "${OUTPUT_DIR_NODEJS18}"
    rm -rf "${OUTPUT_DIR_NODEJS20}"
    rm -rf "${OUTPUT_DIR_JAVA8}"
    rm -rf "${OUTPUT_DIR_JAVA11}"
    rm -rf "${OUTPUT_DIR_JAVA17}"    
    rm -rf "${OUTPUT_DIR_JAVA21}"
  fi

  if ! [[ -f "${OUTPUT_DIR}/build.oci" ]] || \
     ! [[ -f "${OUTPUT_DIR}/run.oci" ]] || \
     ! [[ -f "${OUTPUT_DIR_NODEJS16}/run.oci" ]] || \
     ! [[ -f "${OUTPUT_DIR_NODEJS18}/run.oci" ]] || \
     ! [[ -f "${OUTPUT_DIR_NODEJS20}/run.oci" ]] || \
     ! [[ -f "${OUTPUT_DIR_JAVA8}/run.oci" ]]  || \
     ! [[ -f "${OUTPUT_DIR_JAVA11}/run.oci" ]]  || \
     ! [[ -f "${OUTPUT_DIR_JAVA17}/run.oci" ]]  || \
     ! [[ -f "${OUTPUT_DIR_JAVA21}/run.oci" ]]; then
    util::print::title "Creating stack..."
    "${STACK_DIR}/scripts/create.sh"
  fi

  util::print::title "Setting up local registry"

  if [[ -z "${regport:-}" ]]; then
    regport="5000"
  fi

  util::print::title "Setting up local registry"

  registry_container_id=$(util::tools::setup_local_registry "$regport")

  export REGISTRY_URL="localhost:${regport}"

  pack config experimental true

  tests::run "${registry_container_id}"

  util::tools::cleanup_local_registry "${registry_container_id}"
}

function usage() {
  cat <<-USAGE
test.sh [OPTIONS]

Runs acceptance tests against the stack. Uses the OCI images
${STACK_DIR}/build/build.oci
and
${STACK_DIR}/build/run.oci
and
${STACK_DIR}/build-nodejs-16/run.oci
and
${STACK_DIR}/build-nodejs-18/run.oci
and
${STACK_DIR}/build-nodejs-20/run.oci
and
${STACK_DIR}/build-java-8/run.oci
and
${STACK_DIR}/build-java-11/run.oci
and
${STACK_DIR}/build-java-17/run.oci
and
${STACK_DIR}/build-java-21/run.oci
if they exist. Otherwise, first runs create.sh to create them.

OPTIONS
  --clean          -c  clears contents of stack output directory before running tests
  --regport <port> -p  Local port to use for local registry during tests, defaults to 5000
  --token <token>  -t  Token used to download assets from GitHub (e.g. jam, pack, etc) (optional)
  --help           -h  prints the command usage
USAGE
}

function tools::install() {
  local token
  token="${1}"

  util::tools::jam::install \
    --directory "${STACK_DIR}/.bin" \
    --token "${token}"

  util::tools::pack::install \
    --directory "${STACK_DIR}/.bin" \
    --token "${token}"

  util::tools::skopeo::check
}

function tests::run() {
  local registry_container_id
  registry_container_id="${1}"

  util::print::title "Run Stack Acceptance Tests"

  export CGO_ENABLED=0
  testout=$(mktemp)
  pushd "${STACK_DIR}" > /dev/null
    if GOMAXPROCS="${GOMAXPROCS:-4}" go test -count=1 -timeout 0 ./... -v -run Acceptance | tee "${testout}"; then
      util::tools::tests::checkfocus "${testout}"
      util::tools::cleanup_local_registry "${registry_container_id}"
      util::print::success "** GO Test Succeeded **"
    else
      util::tools::cleanup_local_registry "${registry_container_id}"
      util::print::error "** GO Test Failed **"
    fi
  popd > /dev/null
}


function util::tools::setup_local_registry() {

  registry_port="${1}"

  local registry_container_id
  if [[ "$(curl -s -o /dev/null -w "%{http_code}" localhost:$registry_port/v2/)" == "200" ]]; then
    registry_container_id=""
  else
    registry_container_id=$(docker run -d -p "${registry_port}:5000" --restart=always registry:2)
  fi

  echo $registry_container_id
}

function util::tools::cleanup_local_registry() {
  local registry_container_id
  registry_container_id="${1}"

  if [[ -n "${registry_container_id}" ]]; then
    docker stop $registry_container_id
    docker rm $registry_container_id
  fi
}

main "${@:-}"

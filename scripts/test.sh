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
  local clean regport dont_create_registry
  clean="false"
  regport=""
  dont_create_registry="false"

  while [[ "${#}" != 0 ]]; do
    case "${1}" in
    --help | -h)
      shift 1
      usage
      exit 0
      ;;

    --regport | -r)
      regport="${2}"
      shift 2
      ;;

    --dont-create-registry | -dr)
      dont_create_registry="true"
      shift 1
      ;;

    --clean | -c)
      shift 1
      clean="true"
      ;;

    "")
      # skip if the argument is empty
      shift 1
      ;;

    *)
      util::print::error "unknown argument \"${1}\""
      ;;
    esac
  done

  tools::install

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

  if [[ "${dont_create_registry}" == "true" ]]; then
    echo "Skipping creating local registry"
    tests::run $regport ""
  else
    local local_registry_container_id
    local_registry_container_id=$(docker run -d -p "${regport}:5000" --restart=always --name registry registry:2)
    tests::run $regport $local_registry_container_id
  fi

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
  --clean  -c  clears contents of stack output directory before running tests
  --regport <port> Local port to use for local registry during tests, defaults to 5000
  --dont-create-registry -dr skips creating a local registry, uses the existing one
  --help   -h  prints the command usage
USAGE
}

function tools::install() {
  util::tools::jam::install \
    --directory "${STACK_DIR}/.bin"

  util::tools::pack::install \
    --directory "${STACK_DIR}/.bin"

  util::tools::skopeo::check

  util::tools::docker::check

  util::tools::git::check
}

function tests::run() {
  util::print::title "Run Stack Acceptance Tests"

  local local_registry_port
  local_registry_port="${1}"

  local_registry_container_id="${2}"

  pack config experimental true

  testout=$(mktemp)
  pushd "${STACK_DIR}" >/dev/null
  if GOMAXPROCS="${GOMAXPROCS:-4}" go test -count=1 -timeout 0 ./integration/... -v -run Acceptance --registry-url "localhost:${local_registry_port}" | tee "${testout}"; then
    util::tools::tests::checkfocus "${testout}"
    util::print::success "** GO Test Succeeded **"
  else
    util::print::error "** GO Test Failed **"
  fi

  if [[ -n "${local_registry_container_id}" ]]; then
    docker stop $local_registry_container_id
    docker rm $local_registry_container_id
  fi

  popd >/dev/null
}

main "${@:-}"

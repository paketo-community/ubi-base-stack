#!/usr/bin/env bash

set -eu
set -o pipefail

# shellcheck source=SCRIPTDIR/print.sh
source "$(dirname "${BASH_SOURCE[0]}")/print.sh"

function util::tools::os() {
  case "$(uname)" in
    "Darwin")
      echo "${1:-darwin}"
      ;;

    "Linux")
      echo "linux"
      ;;

    *)
      util::print::error "Unknown OS \"$(uname)\""
      exit 1
  esac
}

function util::tools::arch() {
  case "$(uname -m)" in
    arm64|aarch64)
      echo "arm64"
      ;;

    amd64|x86_64)
      if [[ "${1:-}" == "--blank-amd64" ]]; then
        echo ""
      else
        echo "amd64"
      fi
      ;;

    *)
      util::print::error "Unknown Architecture \"$(uname -m)\""
      exit 1
  esac
}

function util::tools::path::export() {
  local dir
  dir="${1}"

  if ! echo "${PATH}" | grep -q "${dir}"; then
    PATH="${dir}:$PATH"
    export PATH
  fi
}

function util::tools::jam::install() {
  local dir token
  token=""

  while [[ "${#}" != 0 ]]; do
    case "${1}" in
      --directory)
        dir="${2}"
        shift 2
        ;;

      --token)
        token="${2}"
        shift 2
        ;;

      *)
        util::print::error "unknown argument \"${1}\""
    esac
  done

  mkdir -p "${dir}"
  util::tools::path::export "${dir}"

  if [[ ! -f "${dir}/jam" ]]; then
    local version curl_args os arch

    version="$(jq -r .jam "$(dirname "${BASH_SOURCE[0]}")/tools.json")"

    curl_args=(
      "--fail"
      "--silent"
      "--location"
      "--output" "${dir}/jam"
    )

    if [[ "${token}" != "" ]]; then
      curl_args+=("--header" "Authorization: Token ${token}")
    fi

    util::print::title "Installing jam ${version}"

    os=$(util::tools::os)
    arch=$(util::tools::arch)

    curl "https://github.com/paketo-buildpacks/jam/releases/download/${version}/jam-${os}-${arch}" \
      "${curl_args[@]}"

    chmod +x "${dir}/jam"
  else
    util::print::info "Using $("${dir}"/jam version)"
  fi
}

function util::tools::pack::install() {
  local dir token
  token=""

  while [[ "${#}" != 0 ]]; do
    case "${1}" in
      --directory)
        dir="${2}"
        shift 2
        ;;

      --token)
        token="${2}"
        shift 2
        ;;

      *)
        util::print::error "unknown argument \"${1}\""
    esac
  done

  mkdir -p "${dir}"
  util::tools::path::export "${dir}"

  if [[ ! -f "${dir}/pack" ]]; then
    local version curl_args os arch

    version="$(jq -r .pack "$(dirname "${BASH_SOURCE[0]}")/tools.json")"

    tmp_location="/tmp/pack.tgz"
    curl_args=(
      "--fail"
      "--silent"
      "--location"
      "--output" "${tmp_location}"
    )

    if [[ "${token}" != "" ]]; then
      curl_args+=("--header" "Authorization: Token ${token}")
    fi

    util::print::title "Installing pack ${version}"

    os=$(util::tools::os macos)
    arch=$(util::tools::arch --blank-amd64)

    curl "https://github.com/buildpacks/pack/releases/download/${version}/pack-${version}-${os}${arch:+-$arch}.tgz" \
      "${curl_args[@]}"

    tar xzf "${tmp_location}" -C "${dir}"
    chmod +x "${dir}/pack"

    rm "${tmp_location}"
  else
    util::print::info "Using pack $("${dir}"/pack version)"
  fi
}

function util::tools::syft::install() {
  local dir token
  token=""

  while [[ "${#}" != 0 ]]; do
    case "${1}" in
      --directory)
        dir="${2}"
        shift 2
        ;;

      --token)
        token="${2}"
        shift 2
        ;;

      *)
        util::print::error "unknown argument \"${1}\""
    esac
  done

  mkdir -p "${dir}"
  util::tools::path::export "${dir}"

  if [[ ! -f "${dir}/syft" ]]; then
    local version curl_args os arch

    version="$(jq -r .syft "$(dirname "${BASH_SOURCE[0]}")/tools.json")"

    tmp_location="/tmp/syft.tgz"
    curl_args=(
      "--fail"
      "--silent"
      "--location"
      "--output" "${tmp_location}"
    )

    if [[ "${token}" != "" ]]; then
      curl_args+=("--header" "Authorization: Token ${token}")
    fi

    util::print::title "Installing syft ${version}"

    os=$(util::tools::os)
    arch=$(util::tools::arch)

    curl "https://github.com/anchore/syft/releases/download/${version}/syft_${version#v}_${os}_${arch}.tar.gz" \
      "${curl_args[@]}"

    tar xzf "${tmp_location}" -C "${dir}"
    chmod +x "${dir}/syft"

    rm "${tmp_location}"
  fi
}

function util::tools::skopeo::check () {
  if ! command -v  skopeo &> /dev/null ; then
      util::print::error "skopeo could not be found. Please install skopeo before proceeding."
  fi

  local version
  version="v$(skopeo -v | awk '{ print $3}')"

  util::print::title "Using installed skopeo version ${version}"
}

function util::tools::tests::checkfocus() {
  testout="${1}"
  if grep -q 'Focused: [1-9]' "${testout}"; then
    echo "Detected Focused Test(s) - setting exit code to 197"
    rm "${testout}"
    util::print::success "** GO Test Succeeded **" 197
  fi
  rm "${testout}"
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

  util::print::title "inside cleanup $registry_container_id"

  if [[ -n "${registry_container_id}" ]]; then
    docker stop $registry_container_id
    docker rm $registry_container_id
  fi
}
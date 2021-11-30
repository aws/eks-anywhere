#!/usr/bin/env bash

set -x
set -o errexit
set -o nounset
set -o pipefail

ARTIFACTS_BUCKET="${1?Specify first argument - artifacts buckets}"
PROJECT_PATH="${2? Specify second argument - project path}"
BUILD_IDENTIFIER="${3? Specify third argument - build identifier}"
GIT_HASH="${4?Specify fourth argument - git hash of the tar builds}"

REPO="eksctl-anywhere"
BINARY_PATH="bin"
TAR_PATH="${CODEBUILD_SRC_DIR}/${PROJECT_PATH}/${BUILD_IDENTIFIER}-${GIT_HASH}/artifacts"

function build::cli::move_artifacts() {
  local -r os=$1
  local -r arch=$2
  local -r cli_artifacts_path=$3

  mv ${BINARY_PATH}/${os}/${arch}/eksctl-anywhere ${cli_artifacts_path}
  cp ATTRIBUTION.txt ${cli_artifacts_path}
}

function build::cli::create_tarball() {
  local -r os=$1
  local -r arch=$2
  local -r tar_file=$3
  local -r tar_path=$4
  local -r cli_artifacts_path=$5

  build::ensure_tar

  "${TAR}" czf "${tar_path}/${os}/${tar_file}" -C ${cli_artifacts_path} . --owner=0 --group=0
}

function build::cli::generate_shasum() {
  local -r tar_path=$1
  local -r os=$2

  echo "Writing artifact hashes to shasum files..."

  if [ ! -d "$tar_path" ]; then
    echo "  Unable to find tar directory $tar_path"
    exit 1
  fi

  cd $tar_path/$os
  for file in $(find . -name '*.tar.gz'); do
    filepath=$(basename $file)
    sha256sum "$filepath" > "$file.sha256"
    sha512sum "$filepath" > "$file.sha512"
  done
  cd -
}

function build::ensure_tar() {
  if [[ -n "${TAR:-}" ]]; then
    return
  fi

  # Find gnu tar if it is available, bomb out if not.
  TAR=tar
  if which gtar &>/dev/null; then
      TAR=gtar
  elif which gnutar &>/dev/null; then
      TAR=gnutar
  fi
  if ! "${TAR}" --version | grep -q GNU; then
    echo "  !!! Cannot find GNU tar. Build on Linux or install GNU tar"
    echo "      on Mac OS X (brew install gnu-tar)."
    return 1
  fi
}

function build::cli::upload() {
  local -r artifactspath=$1
  local -r artifactsbucket=$2
  local -r projectpath=$3
  local -r buildidentifier=$4
  local -r githash=$5

  echo "$githash" >> "$artifactspath"/githash

  # Upload artifacts to s3
  # 1. To proper path on s3 with buildId-githash
  # 2. Latest path to indicate the latest build, with --delete option to delete stale files in the dest path
  aws s3 sync "$artifactspath" "$artifactsbucket"/"$projectpath"/"$buildidentifier"-"$githash"/artifacts --acl public-read
  aws s3 sync "$artifactspath" "$artifactsbucket"/"$projectpath"/latest --delete --acl public-read
}

SUPPORTED_PLATFORMS=(
  "linux/amd64"
  "darwin/amd64"
)

for platform in "${SUPPORTED_PLATFORMS[@]}"; do
  OS="$(cut -d '/' -f1 <<< ${platform})"
  ARCH="$(cut -d '/' -f2 <<< ${platform})"
  TAR_FILE="${REPO}-${OS}-${ARCH}.tar.gz"
  CLI_ARTIFACTS_PATH="cli-artifacts/${OS}"
  mkdir -p $TAR_PATH/$OS
  mkdir -p $CLI_ARTIFACTS_PATH

  build::cli::move_artifacts $OS $ARCH $CLI_ARTIFACTS_PATH
  build::cli::create_tarball $OS $ARCH $TAR_FILE $TAR_PATH $CLI_ARTIFACTS_PATH
  build::cli::generate_shasum $TAR_PATH $OS
done
build::cli::upload ${TAR_PATH} ${ARTIFACTS_BUCKET} ${PROJECT_PATH} ${BUILD_IDENTIFIER} ${GIT_HASH}

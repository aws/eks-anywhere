#!/usr/bin/env bash

set -x
set -o errexit
set -o nounset
set -o pipefail

ARTIFACTS_BUCKET="${1?Specify first argument - artifacts bucket}"
REPO_ROOT="${2?Specify second argument - repo root directory}"
PROJECT_PATH="${3? Specify third argument - project path}"
BUILD_IDENTIFIER="${4? Specify fourth argument - build identifier}"
GIT_HASH="${5?Specify fifth argument - git hash of the tar builds}"
OS_LIST_CSV="${6?Specify sixth argument - comma-separated list of operating systems for CLI build}"
ARCH_LIST_CSV="${7?Specify fifth argument - comma-separated list of architectures for CLI build}"
DRY_RUN="${8?Specify sixth argument - Dry run upload}"

REPO="eksctl-anywhere"
BINARY_PATH="bin"
TAR_PATH="${REPO_ROOT}/${PROJECT_PATH}/${BUILD_IDENTIFIER}-${GIT_HASH}/artifacts"
LATEST=latest
if [[ $BRANCH_NAME != "main" ]]; then
  LATEST=$BRANCH_NAME
fi

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

  "${TAR}" czf "${tar_path}/${os}/${arch}/${tar_file}" -C ${cli_artifacts_path} . --owner=0 --group=0
}

function build::cli::generate_shasum() {
  local -r tar_path=$1
  local -r os=$2
  local -r arch=$3

  echo "Writing artifact hashes to shasum files..."

  if [ ! -d "$tar_path" ]; then
    echo "  Unable to find tar directory $tar_path"
    exit 1
  fi

  cd $tar_path/$os/$arch
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
  local -r latesttag=$6
  local -r dry_run=$7

  echo "$githash" >> "$artifactspath"/githash

  if [ "$dry_run" = "true" ]; then
    aws s3 cp "$artifactspath" "$artifactsbucket"/"$projectpath"/"$buildidentifier"-"$githash"/artifacts --recursive --dryrun
    aws s3 cp "$artifactspath" "$artifactsbucket"/"$projectpath"/"$latesttag" --recursive --dryrun
  else
    # Upload artifacts to s3
    # 1. To proper path on s3 with buildId-githash
    # 2. Latest path to indicate the latest build, with --delete option to delete stale files in the dest path
    aws s3 sync "$artifactspath" "$artifactsbucket"/"$projectpath"/"$buildidentifier"-"$githash"/artifacts --acl public-read
    aws s3 sync "$artifactspath" "$artifactsbucket"/"$projectpath"/"$latesttag" --delete --acl public-read
  fi
}

OS_LIST=($(echo $OS_LIST_CSV | tr "," "\n"))
ARCH_LIST=($(echo $ARCH_LIST_CSV | tr "," "\n"))

for OS in "${OS_LIST[@]}"; do
  for ARCH in "${ARCH_LIST[@]}"; do
    TAR_FILE="${REPO}-${OS}-${ARCH}.tar.gz"
    CLI_ARTIFACTS_PATH="cli-artifacts/${OS}/${ARCH}"
    mkdir -p $TAR_PATH/$OS/$ARCH
    mkdir -p $CLI_ARTIFACTS_PATH

    build::cli::move_artifacts $OS $ARCH $CLI_ARTIFACTS_PATH
    build::cli::create_tarball $OS $ARCH $TAR_FILE $TAR_PATH $CLI_ARTIFACTS_PATH
    build::cli::generate_shasum $TAR_PATH $OS $ARCH
  done
done
build::cli::upload ${TAR_PATH} ${ARTIFACTS_BUCKET} ${PROJECT_PATH} ${BUILD_IDENTIFIER} ${GIT_HASH} ${LATEST} ${DRY_RUN}

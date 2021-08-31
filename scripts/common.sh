#!/usr/bin/env bash
# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

function build::gather_licenses() {
  if ! command -v go-licenses &> /dev/null
  then
    echo " go-licenses not found.  If you need license or attribtuion file handling"
    echo " please refer to the doc in docs/development/attribution-files.md"
    exit
  fi

  local -r outputdir=$1
  local -r patterns=$2

  # reset the gopath change to make sure and always use
  # the latest go for generating deps list
  # older versions behave differently in some cases
  build::common::remove_go_path

  # Force deps to only be pulled form vendor directories
  # this is important in a couple cases where license files
  # have to be manually created
  export GOFLAGS=-mod=vendor
  # force platform to be linux to ensure all deps are picked up
  export GOOS=linux
  export GOARCH=amd64

  mkdir -p "${outputdir}/attribution"
  # attribution file generated uses the output go-deps and go-license to gather the neccessary
  # data about each dependency to generate the amazon approved attribution.txt files
  # go-deps is needed for module versions
  # go-licenses are all the dependencies found from the module(s) that were passed in via patterns
  go list -deps=true -json ./... | jq -s ''  > "${outputdir}/attribution/go-deps.json"

  go-licenses save --force $patterns --save_path="${outputdir}/LICENSES"

  # go-licenses can be a bit noisy with its output and lot of it can be confusing
  # the following messags are safe to ignore since we do not need the license url for our process
  NOISY_MESSAGES="cannot determine URL for|Error discovering URL|unsupported package host"
  go-licenses csv $patterns > "${outputdir}/attribution/go-license.csv" 2>  >(grep -vE "$NOISY_MESSAGES" >&2)

  if cat "${outputdir}/attribution/go-license.csv" | grep -q "^vendor\/golang.org\/x"; then
      echo " go-licenses created a file with a std golang package (golang.org/x/*)"
      echo " prefixed with vendor/.  This most likely will result in an error"
      echo " when generating the attribution file and is probably due to"
      echo " to a version mismatch between the current version of go "
      echo " and the version of go that was used to build go-licenses"
      exit 1
  fi

  if cat "${outputdir}/attribution/go-license.csv" | grep -e ",LGPL-" -e ",GPL-"; then
    echo " one of the dependencies is licensed as LGPL or GPL"
    echo " which is prohibited at Amazon"
    echo " please look into removing the dependency"
  fi


  # go-license is pretty eager to copy src for certain license types
  # when it does, it applies strange permissions to the copied files
  # which makes deleting them later awkward
  # this behavior may change in the future with the following PR
  # https://github.com/google/go-licenses/pull/28
  chmod -R 777 "${outputdir}/LICENSES"

  # most of the packages show up the go-license.csv file as the module name
  # from the go.mod file, storing that away since the source dirs usually get deleted
  MODULE_NAME=$(go mod edit -json | jq -r '.Module.Path')
  echo $MODULE_NAME > ${outputdir}/attribution/root-module.txt
}

function build::generate_attribution(){
  local -r golang_verson=$1
  local -r output_directory=${2:-"_output"}
  local -r attribution_file=${3:-"ATTRIBUTION.txt"}

  local -r root_module_name=$(cat ${output_directory}/attribution/root-module.txt)
  local -r go_path=$(build::common::get_go_path $golang_verson)
  local -r golang_version_tag=$($go_path/go version | grep -o "go[0-9].* ")

  if cat "${output_directory}/attribution/go-license.csv" | grep -e ",LGPL-" -e ",GPL-"; then
    echo " one of the dependencies is licensed as LGPL or GPL"
    echo " which is prohibited at Amazon"
    echo " please look into removing the dependency"
    exit 1
  fi

  generate-attribution $root_module_name "./" $golang_version_tag $output_directory
  cp -f "${output_directory}/attribution/ATTRIBUTION.txt" $attribution_file
}

function build::common::remove_go_path() {
  # This is the path where the specific go binary versions reside in our builder-base image
  export PATH=$(echo "$PATH" | sed -e "s/^\/go\/go[0-9\.]*\/bin://")

  # This is the path that will most likely be correct if running locally
  local -r quoted_go_path=$(build::common::re_quote $GOPATH)
  export PATH=$(echo "$PATH" | sed -e "s/^$quoted_go_path\/go[0-9\.]*\/bin://")
}

function build::common::get_go_path() {
  local -r version=$1
  local gobinaryversion=""

  if [[ $version == "1.13"* ]]; then
    gobinaryversion="1.13"
  fi
  if [[ $version == "1.14"* ]]; then
    gobinaryversion="1.14"
  fi
  if [[ $version == "1.15"* ]]; then
    gobinaryversion="1.15"
  fi
  if [[ $version == "1.16"* ]]; then
    gobinaryversion="1.16"
  fi

  if [[ "$gobinaryversion" == "" ]]; then
    return
  fi

  # This is the path where the specific go binary versions reside in our builder-base image
  local -r gorootbinarypath="/go/go${gobinaryversion}/bin"
  # This is the path that will most likely be correct if running locally
  local -r gopathbinarypath="$GOPATH/go${gobinaryversion}/bin"
  if [ -d "$gorootbinarypath" ]; then
    echo $gorootbinarypath
  elif [ -d "$gopathbinarypath" ]; then
    echo $gopathbinarypath
  else
    # not in builder-base, probably running in dev environment
    # return default go installation
    local -r which_go=$(which go)
    echo "$(dirname $which_go)"
  fi
}

function build::common::use_go_version() {
  local -r version=$1

  local -r gobinarypath=$(build::common::get_go_path $version)
  echo "Adding $gobinarypath to PATH"
  # Adding to the beginning of PATH to allow for builds on specific version if it exists
  export PATH=${gobinarypath}:$PATH
}

function build::common::re_quote() {
    local -r to_escape=$1
    sed 's/[][()\.^$\/?*+]/\\&/g' <<< "$to_escape"
}

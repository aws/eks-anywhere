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

BUILD_LIB_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/" && pwd -P)"

if [ -f "/buildkit.sh" ] && ! buildctl debug workers > /dev/null 2>&1; then
    # on the builder base this helper file exists to run buildkitd
    # in prow buildkitd is run as a seperate container so it will be running already
    # in codebuild its run directly on the instance so we want to use this helper
    /buildkit.sh "$@"
else
    buildctl "$@"
fi
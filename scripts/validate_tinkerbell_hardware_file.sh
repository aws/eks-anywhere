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

E2E_BINARY="${1?Specify first argument - E2E tests binary for the current build}"
TINKERBELL_HARDWARE_COUNT_FILE="${2?Specify second argument - tinkerbell hardware requirements file}"

ALL_TINKERBELL_TESTS=($($E2E_BINARY -test.list 'TestTinkerbell'))

declare -A TINKERBELL_HARDWARE_COUNT_LIST

HARDWARE_COUNT_VALIDATION_STATUS=true 
while IFS="@" read -r TEST_NAME HARDWARE_COUNT
# create associative array of test name with count
do
    TINKERBELL_HARDWARE_COUNT_LIST[$TEST_NAME]="$HARDWARE_COUNT"
done < <(yq e 'to_entries | .[] | (.key + "@" + .value)' ${TINKERBELL_HARDWARE_COUNT_FILE})
for test in ${ALL_TINKERBELL_TESTS[@]}
do 
    if [[ -n "${TINKERBELL_HARDWARE_COUNT_LIST[$test]}" ]]; then 
        count=${TINKERBELL_HARDWARE_COUNT_LIST[$test]}
    # check if the count value is a number and between 1-10
        if ! [[ $count =~ ^[0-9]+$ ]]; then
            HARDWARE_COUNT_VALIDATION_STATUS=false 
            echo "hardware count for $test is not a integer"
        else
            if ! [[ $count -ge 1 && $count -le 10 ]]; then
                HARDWARE_COUNT_VALIDATION_STATUS=false 
                echo "$test has count higher than permissible range 1 - 10"
            fi
        fi
    # validation fails if any test is missed from the count file
    else
        HARDWARE_COUNT_VALIDATION_STATUS=false
        echo "$test not found in $TINKERBELL_HARDWARE_COUNT_FILE"
    fi
done
if [ $HARDWARE_COUNT_VALIDATION_STATUS = false ]; then
    echo "Hardware Count file validations failed"
    exit 1
else    
    echo "Hardware Count file validations passed!"
    exit 0
fi
#!/usr/bin/env bash

set -x
set -e
set -o pipefail

git config --global credential.helper '!aws codecommit credential-helper $@'
git config --global credential.UseHttpPath true

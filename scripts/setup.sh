#!/usr/bin/env bash

set -x
set -e
set -o pipefail

if [ ! -d "/root/.docker" ]; then
    mkdir -p /root/.docker
fi
mv manager/config/docker-ecr-config.json /root/.docker/config.json
git config --global credential.helper '!aws codecommit credential-helper $@'
git config --global credential.UseHttpPath true

#!/usr/bin/env bash

set -x
set -e
set -o pipefail

if [ ! -d "/root/.docker" ]; then
    mkdir -p /root/.docker
fi
if [ ! -d "/root/.config/containers" ]; then
    mkdir -p /root/.config/containers
fi
mv release/scripts/docker-ecr-config.json /root/.docker/config.json
mv release/scripts/policy.json /root/.config/containers/policy.json
git config --global credential.helper '!aws codecommit credential-helper $@'
git config --global credential.UseHttpPath true

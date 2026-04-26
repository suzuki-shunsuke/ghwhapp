#!/usr/bin/env bash

set -eux

cd "$(dirname "$0")"

gh release download --pattern ghwhapp_linux_arm64.zip

cp config.yaml.tmpl config.yaml
cp secret.yaml.tmpl secret.yaml

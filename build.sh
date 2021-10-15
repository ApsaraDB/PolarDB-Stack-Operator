#!/usr/bin/env bash

APP_NAME=polar-mpd-controller
BUILD_VERSION=0.0.1
BUILD_IMAGE=reg.docker.alibaba-inc.com/polarbox/${APP_NAME}

echo "docker build -t ${BUILD_IMAGE}:${BUILD_VERSION}-SNAPSHOT ."
docker build -t "${BUILD_IMAGE}:${BUILD_VERSION}-SNAPSHOT" \
	-f Dockerfile \
	--build-arg ssh_prv_key="$(cat ~/.ssh/id_rsa)" \
	--build-arg ssh_pub_key="$(cat ~/.ssh/id_rsa.pub)" .

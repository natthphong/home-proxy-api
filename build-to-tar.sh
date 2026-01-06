#!/usr/bin/env bash
set -e

APP_NAME=home-proxy
IMAGE_TAG=3
OUTPUT_TAR=${APP_NAME}-${IMAGE_TAG}.tar
IMAGE_NAME=${APP_NAME}:${IMAGE_TAG}
echo "==> Build Go binary"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o goapp

echo "==> Build docker image: ${IMAGE_NAME}"
docker build -t ${IMAGE_NAME} .

echo "==> Save docker image to tar: ${OUTPUT_TAR}"
docker save -o ${OUTPUT_TAR} ${IMAGE_NAME}

echo "==> Done"
echo "Image: ${IMAGE_NAME}"
echo "Tar  : ${OUTPUT_TAR}"

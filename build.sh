#!/bin/sh

docker buildx build --platform linux/arm64 -t retro-console-arm64 -f Dockerfile .
docker create --name extract retro-console-arm64
docker cp extract:/build/mortar mortar
docker rm extract

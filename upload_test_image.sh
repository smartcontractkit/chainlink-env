#!/usr/bin/env bash
docker build --no-cache -t core-integration-tests:$2 --build-arg HELM_VERSION=3.9.4 --build-arg KUBE_VERSION="v1.25.0" .
docker tag core-integration-tests:$2 $1/core-integration-tests:$2
docker push $1/core-integration-tests:$2
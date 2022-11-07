#!/usr/bin/env bash
docker build -f Dockerfile.base --no-cache -t test-base-image:$2 --build-arg HELM_VERSION=3.9.4 --build-arg KUBE_VERSION="v1.25.0" .
docker tag test-base-image:$2 $1/test-base-image:$2
docker push $1/test-base-image:$2

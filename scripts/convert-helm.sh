#!/bin/bash

mkdir -p ./k8s

helm dependency build ./helm
helm template bubblebank ./helm > ./k8s/manifests.yaml

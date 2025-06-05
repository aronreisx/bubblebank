#!/bin/bash

mkdir -p ./k8s/bubblebank

helm dependency build ./helm/bubblebank
helm template bubblebank ./helm/bubblebank > ./k8s/bubblebank/manifests.yaml

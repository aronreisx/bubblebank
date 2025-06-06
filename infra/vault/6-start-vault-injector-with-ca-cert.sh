#!/bin/bash

# Update the dependencies
helm dependency update ~/Projects/bubblebank/infra/helm/vault-injector

# Install the vault-injector
helm install vault-injector ~/Projects/bubblebank/infra/helm/vault-injector \
  --set caCert.data="$(base64 -i tls/vault-ca.crt)"
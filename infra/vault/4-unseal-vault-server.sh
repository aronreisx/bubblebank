#!/bin/bash

# Set to true for HTTPS, false for HTTP
USE_HTTPS=false

if [ "$USE_HTTPS" = true ]; then
  export VAULT_ADDR="https://localhost:8200"
  export VAULT_CACERT="$(pwd)/tls/vault-ca.crt"
else
  export VAULT_ADDR="http://localhost:8200"
fi

vault operator init -key-shares=1 -key-threshold=1 -format=json > init.json

vault operator unseal $(jq -r '.unseal_keys_b64[0]' init.json)

vault login $(jq -r '.root_token' init.json)
#!/bin/bash

# Start a new pod to run the commands
# kubectl run test-vault --image=curlimages/curl --rm -it --restart=Never -- sh

# Check the connection between you Vault server and the cluster using the Health API
# curl -k https://172.17.0.1:8200/v1/sys/health

# lsof -i :8200
# kill 32217

# Start the Vault server in the background
# With logs
# nohup vault server -config=vault.hcl > vault.log 2>&1 &

# Without logs
nohup vault server -config=vault.hcl > /dev/null 2>&1 &
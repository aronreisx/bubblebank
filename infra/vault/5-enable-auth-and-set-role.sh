#!/bin/bash

# THE TEST IS FAILING IN THIS FILE

# ============================================
# Step 1: Configure the Kubernetes Auth Method
# Get Kubernetes API server URL
KUBE_HOST=$(kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.server}')

# Create a service account for Vault authentication
kubectl create serviceaccount vault-auth -n default

# Create ClusterRoleBinding for token review
kubectl create clusterrolebinding vault-auth-binding \
    --clusterrole=system:auth-delegator \
    --serviceaccount=default:vault-auth

# Get the JWT token (long-lived)
TOKEN_REVIEW_JWT=$(kubectl create token vault-auth --duration=8760h)

# Get the Kubernetes CA certificate
KUBE_CA_CERT=$(kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}' | base64 --decode)

# Set the Vault address and CA certificate
export VAULT_ADDR="https://localhost:8200"
export VAULT_CACERT="$(pwd)/tls/vault-ca.crt"

# Unseal the Vault server
vault operator unseal $(jq -r '.unseal_keys_b64[0]' init.json)

# Login to the Vault server
vault login $(jq -r '.root_token' init.json)

# List currently enabled auth methods
vault auth list

# Enable Kubernetes auth method if not already enabled
vault auth enable kubernetes

# Configure Vault with your Kind cluster details
vault write auth/kubernetes/config \
    token_reviewer_jwt="$TOKEN_REVIEW_JWT" \
    kubernetes_host="$KUBE_HOST" \
    kubernetes_ca_cert="$KUBE_CA_CERT"

# ============================================
# Step 2: Create a Policy
# Create a policy for your BubbleBank application
vault policy write bubblebank-policy - <<EOF
# Allow reading secrets for the application
path "secret/data/bubblebank/*" {
  capabilities = ["read"]
}

# Allow reading database credentials
path "secret/data/database/*" {
  capabilities = ["read"]
}

# Allow token operations
path "auth/token/renew-self" {
  capabilities = ["update"]
}

path "auth/token/lookup-self" {
  capabilities = ["read"]
}
EOF

# ============================================
# Step 3: Create the Kubernetes Auth Role
# Create a role that binds service accounts to the policy
vault write auth/kubernetes/role/bubblebank \
    bound_service_account_names=default \
    bound_service_account_namespaces=default \
    policies=bubblebank-policy \
    ttl=24h \
    max_ttl=24h

# ============================================
# Step 4: Verify the Configuration
# Check the auth method configuration
vault read auth/kubernetes/config

# List available roles
vault list auth/kubernetes/role

# Check the specific role
vault read auth/kubernetes/role/bubblebank


# ============================================
# Test Authentication from Kind Cluster
# Test with a pod using the default service account
# kubectl run vault-test --image=hashicorp/vault:latest --rm -it --restart=Never -- sh

# # Inside the pod:
# export VAULT_ADDR=https://host.docker.internal:8200
# export VAULT_SKIP_VERIFY=true

# # Get the service account JWT
# JWT=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)

# # Authenticate with Vault
# vault write auth/kubernetes/login role=bubblebank jwt="$JWT"

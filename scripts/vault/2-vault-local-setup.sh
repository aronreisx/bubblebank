#!/bin/bash

# 3.2. Write vault.hcl
cat > vault.hcl <<EOF
# ===================================
# Vault Local "Prod-Like" Configuration
# ===================================

storage "file" {
  path = "./data"
}

listener "tcp" {
  address       = "0.0.0.0:8200"
  # tls_cert_file = "tls/vault-server.crt"
  # tls_key_file  = "tls/vault-server.key"
  tls_disable = true
}

ui = true

# Disable mlock on dev machine (but note: in real prod, you WANT mlock)
disable_mlock = true

# Enable Kubernetes auth method at startup (optional: can enable later via CLI)
mount "kubernetes" { type = "kubernetes" }

EOF

#!/bin/bash

# 2.1. Make directories for certs
mkdir -p tls
cd tls

# 2.2. Generate a root CA private key and self-signed CA cert
openssl genrsa -out vault-ca.key 4096
openssl req -x509 -new -nodes -key vault-ca.key \
  -subj "/C=US/ST=Test/L=Test/O=Vault-Local-CA/OU=Dev" \
  -days 3650 -out vault-ca.crt

# 2.3. Generate Vault server private key & CSR (common name = vault.local)
openssl genrsa -out vault-server.key 4096
openssl req -new -key vault-server.key \
  -subj "/C=US/ST=Test/L=Test/O=Vault-Server/OU=Dev/CN=vault.local" \
  -out vault-server.csr

# 2.4. Create a config file for SANs (so “localhost” and “vault.local” both work)
cat > vault-sans.cnf <<EOF
[ v3_req ]
subjectAltName = @alt_names

[ alt_names ]
DNS.1   = vault.local
DNS.2   = localhost
DNS.3   = host.docker.internal # For local development
IP.1    = 127.0.0.1
EOF

# 2.5. Sign the CSR with the CA, adding SANs
openssl x509 -req -in vault-server.csr -CA vault-ca.crt -CAkey vault-ca.key \
  -CAcreateserial -out vault-server.crt -days 3650 \
  -extfile vault-sans.cnf -extensions v3_req

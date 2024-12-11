cd docker/volumes/postgres

certstrap --depot-path certs init --common-name internal-ca --passphrase asdf
certstrap --depot-path certs request-cert --common-name postgresql-server --domain localhost --passphrase zxcv
certstrap --depot-path certs sign postgresql-server --CA internal-ca --passphrase asdf
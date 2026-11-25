guix shell go git nss-certs coreutils -- sh -c 'export SSL_CERT_DIR="$GUIX_ENVIRONMENT/etc/ssl/certs" GIT_SSL_CAINFO="$GUIX_ENVIRONMENT/etc/ssl/certs/ca-certificates.crt" && @INCLUDE:source-unix@'

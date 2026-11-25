# Minimal image: no base image is pulled, so there is nothing to pin or
# patch. It contains only the four static programs and a CA bundle.
# Build it with container/build.sh, which compiles the binaries first.
FROM scratch
COPY ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY numb-form numb-commit numb-date numb-typo /
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt \
    NUMBFORM_DOWNLOADS=/downloads \
    HOME=/downloads
WORKDIR /data
USER 65532:65532
ENTRYPOINT ["/numb-form"]

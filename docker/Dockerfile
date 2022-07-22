FROM @image:alpine@ as certs
# update-ca-certificates builds /etc/ssl/certs/ca-certificates.crt from files in /usr/local/share/ca-certificates
# PANs introduce a chicken-and-egg problem of CA bootstrapping
RUN mkdir -p /usr/local/share/ca-certificates && \
    wget https://pki.palantir.tech/ca-download/PalantirThirdGenRootCA-selfsign.pem -O /usr/local/share/ca-certificates/PalantirThirdGenRootCA-selfsign.pem && \
    cat /usr/local/share/ca-certificates/PalantirThirdGenRootCA-selfsign.pem >> /etc/ssl/certs/ca-certificates.crt && \
    apk --update add ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# the templating is a gödel mechanism. "(InputDistArtifacts Product "linux-amd64")" returns a slice that contains the paths to
# the input dist artifacts for the dister with the name "linux-amd64" for the current product (returned by "Product"). The
# See https://github.com/palantir/godel/wiki/Docker for details.
COPY {{InputBuildArtifact Product "linux-amd64"}} /

STOPSIGNAL SIGINT

ENTRYPOINT ["/sc-metrics"]
CMD ["emit", "--once", "--config", "/secrets/sc-metrics.yml"]

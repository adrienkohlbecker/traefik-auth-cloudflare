FROM golang:1.11

WORKDIR /go/src/github.com/adrienkohlbecker/traefik-auth-cloudflare
COPY . .
# Static build required so that we can safely copy the binary over.
RUN GOARCH=arm64 go install github.com/adrienkohlbecker/traefik-auth-cloudflare

ENTRYPOINT ["/go/bin/linux_arm64/traefik-auth-cloudflare"]

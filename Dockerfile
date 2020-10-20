FROM golang:1.15-alpine as builder

WORKDIR /traefik-auth-cloudflare
COPY . .

RUN go build

###

FROM alpine

# Switch to non-root user
RUN adduser -D myapp
USER myapp
WORKDIR /home/myapp

COPY --from=builder --chown=myapp:myapp /traefik-auth-cloudflare/traefik-auth-cloudflare ./

ENTRYPOINT ["./traefik-auth-cloudflare"]

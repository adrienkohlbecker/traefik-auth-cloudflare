# traefik-auth-cloudflare

Forward auth server to verify Cloudflare Access JWT tokens with traefik

## Description

`traefik-auth-cloudflare` is designed to be a forward auth server for [traefik](https://github.com/containous/traefik) and [Cloudflare Access](https://www.cloudflare.com/products/cloudflare-access/).

When forwarding a user's request to your application, Cloudflare Access will include a signed JWT as a HTTP header. This JWT needs to be authenticated to ensure the request has been signed by Cloudflare and has gone through their servers.

Documentation on how to validate the JWT can be found here https://developers.cloudflare.com/access/setting-up-access/validate-jwt-tokens/.

Using `traefik-auth-cloudflare`, you can configure your `traefik` instance to correctly authenticate cloudflare requests, and you can serve multiple authenticated applications from a single instance.

## Example

Look into the [example](example/) directory to find example `docker-compose.yml` and `traefik.toml` files.

## How to use

- Start an instance of `traefik-auth-cloudflare` in the same docker network as `traefik`. ideally this is a distinct network from your applications.

```bash
# create network for traefik->traefik-auth-cloudflare communication

$ docker network create traefik-auth

# start traefik-auth-cloudflare (default port is 8080)
# you need to set the auth domain you configured on cloudflare

$ docker run -d --network traefik-auth --name traefik-auth-cloudflare akohlbecker/traefik-auth-cloudflare --auth-domain https://foo.cloudflareaccess.com

# add traefik to your `traefik-auth` docker network (left to the reader)

$ docker network connect traefik-auth TRAEFIK_CONTAINER
```

- Configure your router to authenticate requests using `traefik-auth-cloudflare`

```bash
# start your app with auth settings
# the Application Audience (aud) tag needs to be set as an URL parameter: `/auth/{audience}`

$ docker run \
  --label "traefik.http.routers.myapp.middlewares=myapp-auth@docker" \
  --label "traefik.http.middlewares.myapp-auth.forwardauth.address=http://traefik-auth-cloudflare:8080/auth/a83fd537ee93f21e86e51ab3c88f84ef07fd388865c7d0c3236947a8cf79daf5" \
  ....
```

- Optionally, configure traefik to forward the authenticated user header to your application

```bash
# start your app with auth user forward
# the http header is `X-Auth-User`

$ docker run \
  --label "traefik.http.middlewares.myapp-auth.forwardauth.authResponseHeaders=X-Auth-User" \
  ....
```

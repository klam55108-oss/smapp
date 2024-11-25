FROM alpine:3
RUN apk add --update git && \
    git clone https://github.com/avoronoi/traefik-jwt-plugin /plugins-local/src/github.com/avoronoi/traefik-jwt-plugin \
      --depth 1 --single-branch --branch pass-pem-file

FROM traefik:3.2.0
COPY --from=0 /plugins-local /plugins-local
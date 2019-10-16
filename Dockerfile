FROM golang:1.13.1-alpine3.10 as go-build

RUN apk add --update --no-cache make git
RUN adduser -D go-build
WORKDIR /home/go-build
COPY --chown=go-build:go-build . .

USER go-build
RUN go mod verify
RUN XBUILD_ARGS="-static" make bin/linux/amd64/acmeproxy

FROM alpine:3.10 as run
RUN adduser -D acmeproxy
WORKDIR /home/acmeproxy
COPY --from=go-build --chown=acmeproxy:acmeproxy /home/go-build/bin/linux/amd64/acmeproxy acmeproxy

USER acmeproxy
ENV ACMEPROXY_ACME_DIRECTORY_URL "https://acme-v02.api.letsencrypt.org/directory"
ENV ACMEPROXY_HTTP_API_ADDR ":80"

ENTRYPOINT ["./acmeproxy"]
CMD ["serve"]

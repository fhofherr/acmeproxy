#!/usr/bin/env bash

set -e

: "${ACMEPROXY_ACME_DIRECTORY_URL:=https://localhost:14000/dir}"
: "${ACMEPROXY_HTTP_API_ADDR:=localhost:5002}"
: "${ACMEPROXY_IMAGE_TAG:=acmeproxy:latest}"
: "${ACMEPROXY_PEBBLE_DIR:=$PWD/.pebble}"
: "${ACMEPROXY_POD_NAME:=acmeproxy-test-env}"

ACMEPROY_HTTP_API_PORT=$(echo $ACMEPROXY_HTTP_API_ADDR | cut -d':' -f2)

if [ $EUID = 0 ]; then
    echo "You can't be root"
    exit 1
fi

if ! command -v podman > /dev/null 2>&1; then
    echo "You don't have podman installed"
    exit 1
fi

if [ ! -d "$PWD/.git" ] || [ ! -f "$PWD/$0" ]; then
    echo "$0 needs to be called from the project root."
    exit 1
fi

if  ! command podman image exists "$ACMEPROXY_IMAGE_TAG"; then
    echo "$ACMEPROXY_IMAGE_TAG could not be found in local storage"
    exit 1
fi

if [ ! -e "$ACMEPROXY_PEBBLE_DIR" ]; then
    make pebble
fi

function start_test_env {
    command podman pod create \
        --name "$ACMEPROXY_POD_NAME" \
        --publish "$ACMEPROY_HTTP_API_PORT"

    command podman run \
        --pod "$ACMEPROXY_POD_NAME" \
        --detach \
        --env PEBBLE_VA_NOSLEEP=1 \
        letsencrypt/pebble:latest \
        pebble \
        -config /test/config/pebble-config.json \
        -strict \
        -dnsserver localhost:8053

    command podman run \
        --pod "$ACMEPROXY_POD_NAME" \
        --detach \
        letsencrypt/pebble-challtestsrv:latest \
        pebble-challtestsrv \
        -defaultIPv6 "" \
        -defaultIPv4 "127.0.0.1" \
        -http01 "" \
        -https01 ""

    command podman run \
        --pod "$ACMEPROXY_POD_NAME" \
        --detach \
        --volume "$ACMEPROXY_PEBBLE_DIR/test/certs:/tmp/certs" \
        --env LEGO_CA_CERTIFICATES="/tmp/certs/pebble.minica.pem" \
        --env ACMEPROXY_ACME_DIRECTORY_URL="$ACMEPROXY_ACME_DIRECTORY_URL" \
        --env ACMEPROXY_HTTP_API_ADDR="$ACMEPROXY_HTTP_API_ADDR" \
        "$ACMEPROXY_IMAGE_TAG"
}

function remove_test_env {
    command podman pod rm -f "$ACMEPROXY_POD_NAME"
}

case $1 in
    start)
        start_test_env
        ;;
    stop|rm)
        remove_test_env
        ;;
    *)
        echo "Usage: $0 <start|stop|rm>"
esac


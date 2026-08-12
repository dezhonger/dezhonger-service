#!/bin/sh
set -eu

if [ -z "${PUBLIC_DOMAIN:-}" ]; then
  printf '%s\n' 'PUBLIC_DOMAIN is required' >&2
  exit 1
fi

content_certificate_name="${CONTENT_CERT_NAME:-dezhonger-content}"
staging_flag=""
if [ "${STAGING:-0}" = "1" ]; then
  content_certificate_name="${content_certificate_name}-staging"
  staging_flag="--staging"
fi

docker compose run --rm --entrypoint certbot certbot certonly \
  --non-interactive \
  --agree-tos \
  --register-unsafely-without-email \
  --webroot \
  --webroot-path /var/www/certbot \
  --cert-name "$content_certificate_name" \
  -d "knowledge.$PUBLIC_DOMAIN" \
  -d "docs.$PUBLIC_DOMAIN" \
  -d "zmq.$PUBLIC_DOMAIN" \
  -d "rby.$PUBLIC_DOMAIN" \
  -d "math.$PUBLIC_DOMAIN" \
  -d "algo.$PUBLIC_DOMAIN" \
  -d "guwen.$PUBLIC_DOMAIN" \
  $staging_flag

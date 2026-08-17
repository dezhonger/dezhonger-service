#!/bin/sh
set -eu

if [ -z "${PUBLIC_DOMAIN:-}" ]; then
  printf '%s\n' 'PUBLIC_DOMAIN is required' >&2
  exit 1
fi

content_certificate_name="${CONTENT_CERT_NAME:-dezhonger-content}"
staging_flag=""
expand_flag="--expand"
if [ "${STAGING:-0}" = "1" ]; then
  content_certificate_name="${content_certificate_name}-staging"
  staging_flag="--staging"
  expand_flag=""
fi

docker compose run --rm --entrypoint certbot certbot certonly \
  --non-interactive \
  --agree-tos \
  --register-unsafely-without-email \
  --webroot \
  --webroot-path /var/www/certbot \
  --cert-name "$content_certificate_name" \
  $expand_flag \
  -d "knowledge.$PUBLIC_DOMAIN" \
  -d "docs.$PUBLIC_DOMAIN" \
  -d "zmq.$PUBLIC_DOMAIN" \
  -d "rby.$PUBLIC_DOMAIN" \
  -d "math.$PUBLIC_DOMAIN" \
  -d "algo.$PUBLIC_DOMAIN" \
  -d "guwen.$PUBLIC_DOMAIN" \
  -d "english.$PUBLIC_DOMAIN" \
  -d "biology.$PUBLIC_DOMAIN" \
  -d "geography.$PUBLIC_DOMAIN" \
  -d "physics.$PUBLIC_DOMAIN" \
  -d "chemistry.$PUBLIC_DOMAIN" \
  -d "history.$PUBLIC_DOMAIN" \
  -d "puzzle.$PUBLIC_DOMAIN" \
  $staging_flag

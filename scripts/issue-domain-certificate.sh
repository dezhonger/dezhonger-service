#!/bin/sh
set -eu

if [ -z "${PUBLIC_DOMAIN:-}" ]; then
  printf '%s\n' 'PUBLIC_DOMAIN is required' >&2
  exit 1
fi

certificate_name="$PUBLIC_DOMAIN"
staging_flag=""
if [ "${STAGING:-0}" = "1" ]; then
  certificate_name="${PUBLIC_DOMAIN}-staging"
  staging_flag="--staging"
fi

docker compose run --rm --entrypoint certbot certbot certonly \
  --non-interactive \
  --agree-tos \
  --register-unsafely-without-email \
  --webroot \
  --webroot-path /var/www/certbot \
  --cert-name "$certificate_name" \
  -d "$PUBLIC_DOMAIN" \
  -d "www.$PUBLIC_DOMAIN" \
  $staging_flag

#!/bin/sh
set -eu

if [ -z "${PUBLIC_IP:-}" ]; then
  printf '%s\n' 'PUBLIC_IP is required' >&2
  exit 1
fi

certificate_name="$PUBLIC_IP"
staging_flag=""
if [ "${STAGING:-0}" = "1" ]; then
  certificate_name="${PUBLIC_IP}-staging"
  staging_flag="--staging"
fi

docker compose run --rm --entrypoint certbot certbot certonly \
  --non-interactive \
  --agree-tos \
  --register-unsafely-without-email \
  --preferred-profile shortlived \
  --webroot \
  --webroot-path /var/www/certbot \
  --ip-address "$PUBLIC_IP" \
  --cert-name "$certificate_name" \
  $staging_flag

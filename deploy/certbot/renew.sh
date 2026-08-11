#!/bin/sh
set -eu

while true; do
  certbot renew --non-interactive --quiet
  sleep 21600
done

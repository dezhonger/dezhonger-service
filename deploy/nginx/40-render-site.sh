#!/bin/sh
set -eu

certificate_path="/etc/letsencrypt/live/${PUBLIC_IP}/fullchain.pem"
runtime_config="/etc/nginx/runtime/site.conf"

certificate_signature() {
  if [ -s "$certificate_path" ]; then
    sha256sum "$certificate_path" | cut -d ' ' -f 1
  else
    printf '%s' absent
  fi
}

render_site() {
  if [ -s "$certificate_path" ]; then
    envsubst '${PUBLIC_IP}' \
      < /etc/nginx/site-templates/site-tls.conf.template \
      > "$runtime_config"
  else
    cp /etc/nginx/site-templates/site-http.conf.template "$runtime_config"
  fi
}

render_site
initial_signature="$(certificate_signature)"

(
  previous_signature="$initial_signature"
  while sleep 300; do
    current_signature="$(certificate_signature)"
    if [ "$current_signature" != "$previous_signature" ]; then
      render_site
      nginx -s reload || true
      previous_signature="$current_signature"
    fi
  done
) &

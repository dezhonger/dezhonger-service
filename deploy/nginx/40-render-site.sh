#!/bin/sh
set -eu

domain_certificate_path="/etc/letsencrypt/live/${PUBLIC_DOMAIN}/fullchain.pem"
ip_certificate_path="/etc/letsencrypt/live/${PUBLIC_IP}/fullchain.pem"
runtime_config="/etc/nginx/runtime/site.conf"
runtime_ip_config="/etc/nginx/runtime/ip-redirect.conf"

certificate_signature() {
  for certificate_path in "$domain_certificate_path" "$ip_certificate_path"; do
    if [ -s "$certificate_path" ]; then
      sha256sum "$certificate_path" | cut -d ' ' -f 1
    else
      printf '%s\n' absent
    fi
  done
}

render_site() {
  rm -f "$runtime_ip_config"

  if [ -s "$domain_certificate_path" ]; then
    envsubst '${PUBLIC_DOMAIN}' \
      < /etc/nginx/site-templates/site-domain-tls.conf.template \
      > "$runtime_config"

    if [ -s "$ip_certificate_path" ]; then
      envsubst '${PUBLIC_DOMAIN} ${PUBLIC_IP}' \
        < /etc/nginx/site-templates/site-ip-redirect.conf.template \
        > "$runtime_ip_config"
    fi
  elif [ -s "$ip_certificate_path" ]; then
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

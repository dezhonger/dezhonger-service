# Keep the running main website unchanged while updating the blog proxy rules.
ARG NGINX_BASE=dezhonger-service-nginx:latest
FROM ${NGINX_BASE}
COPY deploy/nginx/nginx.conf /etc/nginx/nginx.conf
COPY deploy/nginx/site-content-tls.conf.template /etc/nginx/site-templates/site-content-tls.conf.template

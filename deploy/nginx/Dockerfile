FROM node:24-alpine AS web-builder

WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM nginx:1.29-alpine
COPY deploy/nginx/nginx.conf /etc/nginx/nginx.conf
COPY deploy/nginx/site-http.conf.template /etc/nginx/site-templates/site-http.conf.template
COPY deploy/nginx/site-tls.conf.template /etc/nginx/site-templates/site-tls.conf.template
COPY deploy/nginx/site-domain-tls.conf.template /etc/nginx/site-templates/site-domain-tls.conf.template
COPY deploy/nginx/site-ip-redirect.conf.template /etc/nginx/site-templates/site-ip-redirect.conf.template
COPY deploy/nginx/40-render-site.sh /docker-entrypoint.d/40-render-site.sh
COPY --from=web-builder /src/web/dist/ /usr/share/nginx/html/
RUN chmod +x /docker-entrypoint.d/40-render-site.sh \
    && mkdir -p /etc/nginx/runtime /var/www/certbot

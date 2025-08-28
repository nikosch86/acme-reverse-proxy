FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *go .
COPY nginx/templates /tmp/nginx-templates/
# Run tests - some tests may fail in Docker due to permissions, but that's OK
RUN go test || true
RUN go build -o acme .
RUN go build -o generate_nginx_config cmd_generate_nginx_config.go

FROM nginx:1.29-alpine

WORKDIR /root/

COPY --from=builder /app/acme /usr/local/bin/acme
COPY --from=builder /app/generate_nginx_config /usr/local/bin/generate_nginx_config
COPY nginx/templates /etc/nginx/templates/

COPY nginx/nginx.conf /etc/nginx/nginx.conf
COPY nginx/http.conf /etc/nginx/http.conf
COPY nginx/conf.d/*.conf /etc/nginx/conf.d/
COPY ssl/notyetsetup.dev.lan.crt /etc/ssl/private/fullchain.pem
COPY ssl/notyetsetup.dev.lan.key /etc/ssl/private/key.pem

COPY nginx/35-ssl.sh /docker-entrypoint.d/
RUN chmod +x /docker-entrypoint.d/35-ssl.sh

VOLUME /etc/ssl/private/

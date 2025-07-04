FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go test -v ./...
RUN go build -o acme acme.go services.go
RUN go build -o generate_nginx_config generate_nginx_config.go services.go

FROM nginx:1.27-alpine

WORKDIR /root/

COPY --from=builder /app/acme /usr/local/bin/acme
COPY --from=builder /app/generate_nginx_config /usr/local/bin/generate_nginx_config

COPY nginx/nginx.conf /etc/nginx/nginx.conf
COPY nginx/http.conf /etc/nginx/http.conf
COPY nginx/conf.d/*.conf /etc/nginx/conf.d/
COPY ssl/notyetsetup.dev.lan.crt /etc/ssl/private/fullchain.pem
COPY ssl/notyetsetup.dev.lan.key /etc/ssl/private/key.pem

COPY nginx/35-ssl.sh /docker-entrypoint.d/
RUN chmod +x /docker-entrypoint.d/35-ssl.sh

VOLUME /etc/ssl/private/

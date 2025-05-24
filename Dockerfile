FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *go .
RUN go test
RUN go build -o acme .

FROM nginx:1.27-alpine

WORKDIR /root/

COPY --from=builder /app/acme /usr/local/bin/acme

COPY nginx/nginx.conf /etc/nginx/nginx.conf
COPY nginx/http.conf /etc/nginx/http.conf
COPY nginx/conf.d/*.conf /etc/nginx/conf.d/
COPY ssl/notyetsetup.dev.lan.crt /etc/ssl/private/fullchain.pem
COPY ssl/notyetsetup.dev.lan.key /etc/ssl/private/key.pem

COPY nginx/35-ssl.sh /docker-entrypoint.d/
RUN chmod +x /docker-entrypoint.d/35-ssl.sh

VOLUME /etc/ssl/private/

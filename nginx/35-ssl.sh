#!/bin/sh

SERVICE=${SERVICE:-service}
PORT=${PORT:-80}
RENEWAL_SECONDS=${RENEWAL_SECONDS:-86400}

if [ -z "$NO_HTTP_SERVICE" ]; then
    # Replace __SERVICE__ with the value of SERVICE
    sed -i "s/__SERVICE__/$SERVICE/g" /etc/nginx/conf.d/reverse-proxy.conf

    # Replace __PORT__ with the value of PORT
    sed -i "s/__PORT__/$PORT/g" /etc/nginx/conf.d/reverse-proxy.conf
else
    # Remove the default http reverse-proxy config
    rm -f /etc/nginx/conf.d/reverse-proxy.conf
fi


run_acme() {
    /usr/local/bin/acme
}

# Run ACME tool after Nginx starts
(
    # Wait for Nginx to start
    sleep 5
    run_acme
) &

(
    # Run ACME tool every RENEWAL_SECONDS seconds
    while true; do
        sleep $RENEWAL_SECONDS
        run_acme
    done
) &

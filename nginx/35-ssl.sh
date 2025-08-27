#!/bin/sh

SERVICE=${SERVICE:-service}
PORT=${PORT:-80}
RENEWAL_SECONDS=${RENEWAL_SECONDS:-86400}
ENABLE_WEBSOCKET=${ENABLE_WEBSOCKET:-false}

if [ -z "$NO_HTTP_SERVICE" ]; then
    # Replace __SERVICE__ with the value of SERVICE
    sed -i "s/__SERVICE__/$SERVICE/g" /etc/nginx/conf.d/reverse-proxy.conf

    # Replace __PORT__ with the value of PORT
    sed -i "s/__PORT__/$PORT/g" /etc/nginx/conf.d/reverse-proxy.conf

    # Configure WebSocket support if enabled
    if [ "$ENABLE_WEBSOCKET" = "true" ]; then
        # Add WebSocket configuration using a more reliable approach
        cat /etc/nginx/conf.d/reverse-proxy.conf | awk '
        /# WEBSOCKET_CONFIG_PLACEHOLDER/ {
            print ""
            print "      # WebSocket support"
            print "      proxy_http_version 1.1;"
            print "      proxy_set_header Upgrade $http_upgrade;"
            print "      proxy_set_header Connection $connection_upgrade;"
            next
        }
        { print }
        ' > /tmp/reverse-proxy.conf.tmp
        mv /tmp/reverse-proxy.conf.tmp /etc/nginx/conf.d/reverse-proxy.conf
    else
        # Remove the WebSocket placeholder comment
        sed -i "/# WEBSOCKET_CONFIG_PLACEHOLDER/d" /etc/nginx/conf.d/reverse-proxy.conf
    fi
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

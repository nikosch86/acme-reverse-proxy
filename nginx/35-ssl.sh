#!/bin/sh

RENEWAL_SECONDS=${RENEWAL_SECONDS:-86400}

if [ -z "$NO_HTTP_SERVICE" ]; then
    # Generate nginx configuration from template
    OUTPUT_FILE=/etc/nginx/conf.d/reverse-proxy.conf /usr/local/bin/generate_nginx_config
    if [ $? -ne 0 ]; then
        echo "Failed to generate nginx configuration"
        exit 1
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

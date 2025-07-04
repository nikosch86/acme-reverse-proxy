#!/bin/sh

RENEWAL_SECONDS=${RENEWAL_SECONDS:-86400}

# Generate nginx configuration based on services
/usr/local/bin/generate_nginx_config


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

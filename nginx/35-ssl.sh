#!/bin/sh

RENEWAL_SECONDS=${RENEWAL_SECONDS:-86400}

# Create auth directory
mkdir -p /etc/nginx/auth

# Function to generate htpasswd file from environment variable
generate_htpasswd() {
    local auth_users="$1"
    local output_file="$2"
    
    if [ -z "$auth_users" ]; then
        return 0
    fi
    
    echo "Generating htpasswd file: $output_file"
    
    # Clear the file first
    > "$output_file"
    
    # Split on comma and process each user:password/hash pair
    IFS=','
    for user_entry in $auth_users; do
        # Split user:password
        user=$(echo "$user_entry" | cut -d: -f1)
        pass=$(echo "$user_entry" | cut -d: -f2-)
        
        if [ -z "$user" ] || [ -z "$pass" ]; then
            echo "Warning: Invalid auth entry (skipping): $user_entry"
            continue
        fi
        
        # Check if password is already hashed (starts with $2y$ or $2a$ or $2b$ for bcrypt)
        if echo "$pass" | grep -q '^\$2[yab]\$'; then
            # Already hashed, use as-is
            echo "$user:$pass" >> "$output_file"
            echo "  Added user '$user' with existing hash"
        else
            # Plaintext password - hash it using htpasswd
            if command -v htpasswd >/dev/null 2>&1; then
                # Use htpasswd to generate bcrypt hash
                htpasswd -nbB "$user" "$pass" >> "$output_file"
                echo "  Added user '$user' with generated hash"
            else
                echo "ERROR: htpasswd not found. Cannot hash plaintext password for user '$user'"
                echo "       Please provide pre-hashed passwords or install apache2-utils"
                exit 1
            fi
        fi
    done
    unset IFS
    
    chmod 644 "$output_file"
}

# Generate global auth file if configured
if [ -n "$BASIC_AUTH_USERS" ]; then
    generate_htpasswd "$BASIC_AUTH_USERS" "/etc/nginx/auth/global.htpasswd"
fi

# Generate service-specific auth files for multi-service mode
for i in $(seq 1 100); do
    auth_var="BASIC_AUTH_SERVICE_$i"
    auth_value=$(eval echo \$$auth_var)
    if [ -n "$auth_value" ]; then
        generate_htpasswd "$auth_value" "/etc/nginx/auth/service_$i.htpasswd"
    fi
done

# Generate single service auth file if configured
if [ -n "$BASIC_AUTH_SERVICE" ]; then
    generate_htpasswd "$BASIC_AUTH_SERVICE" "/etc/nginx/auth/service.htpasswd"
fi

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

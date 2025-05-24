# nginx based reverse proxy with automatic ACME certificate generation using our in-house PKI

See sample `docker-compose.yml` for an example of how to use this image.  
Basically specify the `SERVICE` and the `DOMAIN` as environment variables.

The certificate is checked every 24 hours for expiration and validity.

For advanced use cases a custom config file can be mounted to `/etc/nginx/conf.d/reverse-proxy.conf`.  
To specify a bunch of default best practice options and the certificate itself, include the `ssl.conf` in your config file like so: `include /etc/nginx/conf.d/ssl.conf;`  

The following variables are available for configuration:

- `EXPIRY_DAYS_THRESHOLD` - the threshold for the certificate expiry in days, defaults to 30
- `EMAIL` - the email address to be used for ACME Account registration, this is generally not important as we don't use it anywhere else, defaults to admin@dev.lan
- `SERVICE` - the service name to be used in the reverse proxy configuration, defaults to 'service'
- `PORT` - the port to be used in the reverse proxy configuration, defaults to 80
- `DOMAIN` - the domain name to be used in the reverse proxy configuration, this is mandatory
- `SAN` - the subject alternative name to be used in the reverse proxy configuration, defaults to "", accepts comma seperated value
- `CERT_PATH` - the path to the certificate file, defaults to /etc/ssl/private/fullchain.pem
- `KEY_PATH` - the path to the private key file, defaults to /etc/ssl/private/key.pem
- `CA_DIR_URL` - the URL to the CA directory, defaults to https://acme-staging-v02.api.letsencrypt.org/directory
    The production URL is https://acme-v02.api.letsencrypt.org/directory
- `RENEWAL_SECONDS` - the renewal interval in seconds, defaults to 86400 (24 hours)


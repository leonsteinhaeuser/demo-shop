#!/bin/sh

echo "Starting Demo Shop Frontend..."
echo "Generating configuration from environment variables..."

# Set default values
GATEWAY_SERVICE_URL=${GATEWAY_SERVICE_URL:-"http://localhost:8081"}
GATEWAY_SERVICE_URL=${GATEWAY_SERVICE_URL:-"http://localhost:8084"}
GATEWAY_SERVICE_URL=${GATEWAY_SERVICE_URL:-"http://localhost:8082"}
GATEWAY_SERVICE_URL=${GATEWAY_SERVICE_URL:-"http://localhost:8085"}
GATEWAY_SERVICE_URL=${GATEWAY_SERVICE_URL:-"http://localhost:8083"}

echo "Items Service URL: $GATEWAY_SERVICE_URL"
echo "Users Service URL: $GATEWAY_SERVICE_URL"
echo "Carts Service URL: $GATEWAY_SERVICE_URL"
echo "Checkouts Service URL: $GATEWAY_SERVICE_URL"
echo "Cart Presentation Service URL: $GATEWAY_SERVICE_URL"

# Generate config.js from template
envsubst < /usr/share/nginx/html/config.template.js > /usr/share/nginx/html/js/config.js
echo "Configuration generated successfully!"

# Copy shop.html as index.html for the main shop experience
cp /usr/share/nginx/html/shop.html /usr/share/nginx/html/index.html
echo "Shop page set as default!"

exec nginx -g "daemon off;"

#!/bin/sh

echo "Starting Demo Shop Frontend..."
echo "Generating configuration from environment variables..."

# Set default values
ITEMS_SERVICE_URL=${ITEMS_SERVICE_URL:-"http://localhost:8081"}
USERS_SERVICE_URL=${USERS_SERVICE_URL:-"http://localhost:8084"}
CARTS_SERVICE_URL=${CARTS_SERVICE_URL:-"http://localhost:8082"}
CHECKOUTS_SERVICE_URL=${CHECKOUTS_SERVICE_URL:-"http://localhost:8085"}
CART_PRESENTATION_SERVICE_URL=${CART_PRESENTATION_SERVICE_URL:-"http://localhost:8083"}

echo "Items Service URL: $ITEMS_SERVICE_URL"
echo "Users Service URL: $USERS_SERVICE_URL"
echo "Carts Service URL: $CARTS_SERVICE_URL"
echo "Checkouts Service URL: $CHECKOUTS_SERVICE_URL"
echo "Cart Presentation Service URL: $CART_PRESENTATION_SERVICE_URL"

# Generate config.js from template
envsubst < /usr/share/nginx/html/config.template.js > /usr/share/nginx/html/js/config.js
echo "Configuration generated successfully!"

exec nginx -g "daemon off;"

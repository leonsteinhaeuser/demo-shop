#!/bin/bash

# Health check script for the frontend application
# This script can be used for more complex health checks

HEALTH_STATUS=0
ERRORS=""

# Check if main HTML files exist
if [ ! -f "/usr/share/nginx/html/shop.html" ]; then
    HEALTH_STATUS=1
    ERRORS="${ERRORS}Missing shop.html; "
fi

if [ ! -f "/usr/share/nginx/html/dashboard.html" ]; then
    HEALTH_STATUS=1
    ERRORS="${ERRORS}Missing dashboard.html; "
fi

# Check if essential JS files exist
if [ ! -f "/usr/share/nginx/html/js/config.js" ]; then
    HEALTH_STATUS=1
    ERRORS="${ERRORS}Missing config.js; "
fi

# Check if nginx is running
if ! pgrep nginx > /dev/null; then
    HEALTH_STATUS=1
    ERRORS="${ERRORS}Nginx not running; "
fi

# Output result
if [ $HEALTH_STATUS -eq 0 ]; then
    echo "healthy"
    exit 0
else
    echo "unhealthy: $ERRORS"
    exit 1
fi

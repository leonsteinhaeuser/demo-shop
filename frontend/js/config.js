// Default configuration for development
// This file will be overwritten in production by the generate-config.sh script
window.API_CONFIG = {
  ITEMS_SERVICE_URL: 'http://localhost:8081',
  USERS_SERVICE_URL: 'http://localhost:8084',
  CARTS_SERVICE_URL: 'http://localhost:8082',
  CHECKOUTS_SERVICE_URL: 'http://localhost:8085',
  CART_PRESENTATION_SERVICE_URL: 'http://localhost:8083'
};

console.log('Frontend configuration loaded:', window.API_CONFIG);

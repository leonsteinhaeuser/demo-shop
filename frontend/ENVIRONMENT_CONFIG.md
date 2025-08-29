# Environment Configuration

The frontend service can be configured for different environments using environment variables.

## Environment Variables

The following environment variables configure the frontend API endpoints:

| Variable | Description | Default Value |
|----------|-------------|---------------|
| `ITEMS_SERVICE_URL` | Items service base URL | `http://localhost:8081` |
| `USERS_SERVICE_URL` | Users service base URL | `http://localhost:8084` |
| `CARTS_SERVICE_URL` | Carts service base URL | `http://localhost:8082` |
| `CHECKOUTS_SERVICE_URL` | Checkouts service base URL | `http://localhost:8085` |
| `CART_PRESENTATION_SERVICE_URL` | Cart Presentation service base URL | `http://localhost:8083` |

## Configuration Examples

### Local Development
The default configuration works for local development when services are running on localhost:

```javascript
window.API_CONFIG = {
  ITEMS_SERVICE_URL: 'http://localhost:8081',
  USERS_SERVICE_URL: 'http://localhost:8084',
  CARTS_SERVICE_URL: 'http://localhost:8082',
  CHECKOUTS_SERVICE_URL: 'http://localhost:8085',
  CART_PRESENTATION_SERVICE_URL: 'http://localhost:8083'
};
```

### Docker Compose (External Access)
In the compose.yaml, the frontend is configured to access services via external ports:

```yaml
environment:
  - ITEMS_SERVICE_URL=http://localhost:8081
  - USERS_SERVICE_URL=http://localhost:8084
  - CARTS_SERVICE_URL=http://localhost:8082
  - CHECKOUTS_SERVICE_URL=http://localhost:8085
  - CART_PRESENTATION_SERVICE_URL=http://localhost:8083
```

### Production Environment
For production, you might use different hostnames or internal service names:

```yaml
environment:
  - ITEMS_SERVICE_URL=https://api.example.com/items
  - USERS_SERVICE_URL=https://api.example.com/users
  - CARTS_SERVICE_URL=https://api.example.com/carts
  - CHECKOUTS_SERVICE_URL=https://api.example.com/checkouts
  - CART_PRESENTATION_SERVICE_URL=https://api.example.com/cart-presentation
```

### Kubernetes Environment
For internal cluster communication:

```yaml
environment:
  - ITEMS_SERVICE_URL=http://items-service:8080
  - USERS_SERVICE_URL=http://users-service:8080
  - CARTS_SERVICE_URL=http://carts-service:8080
  - CHECKOUTS_SERVICE_URL=http://checkouts-service:8080
  - CART_PRESENTATION_SERVICE_URL=http://cart-presentation-service:8080
```

## How it Works

1. **Build Time**: The Containerfile copies all frontend files including a `config.template.js` template file
2. **Runtime**: The `generate-config.sh` script runs at container startup and:
   - Reads environment variables
   - Substitutes them into the template using `envsubst`
   - Generates the final `js/config.js` file
3. **Frontend**: The JavaScript code loads the configuration and uses it to build API URLs

This approach allows the same container image to be used across different environments by simply changing the environment variables.

## API Endpoints

The frontend automatically appends the correct API paths to each service URL:

- Items: `${ITEMS_SERVICE_URL}/api/v1/core/items`
- Users: `${USERS_SERVICE_URL}/api/v1/core/users`
- Carts: `${CARTS_SERVICE_URL}/api/v1/core/carts`
- Checkouts: `${CHECKOUTS_SERVICE_URL}/api/v1/core/checkouts`
- Cart Presentation: `${CART_PRESENTATION_SERVICE_URL}/api/v1/presentation/cart`

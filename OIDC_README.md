# Demo Shop OIDC Server

This is a complete OIDC (OpenID Connect) server implementation for the Demo Shop project, built using the [Zitadel OIDC library](https://github.com/zitadel/oidc).

## Features

- Full OIDC compliance with discovery endpoint
- Authorization Code flow with PKCE support
- JWT tokens (RS256 signing)
- User authentication and authorization
- Client credentials management
- Token introspection and revocation
- Custom login UI
- In-memory storage (demo purposes)

## Getting Started

### Running the OIDC Server

```bash
# Build and run the OIDC server
go run cmd/oidc/main.go

# Or build and run as executable
go build -o oidc-server cmd/oidc/main.go
./oidc-server
```

The server will start on `http://localhost:8080`

### Endpoints

#### OIDC Discovery

- **Discovery**: `GET /api/v1/auth/oidc/.well-known/openid_configuration`

#### OIDC Standard Endpoints

- **Authorization**: `GET/POST /api/v1/auth/oidc/auth`
- **Token**: `POST /api/v1/auth/oidc/token`
- **UserInfo**: `GET/POST /api/v1/auth/oidc/userinfo`
- **Keys (JWKS)**: `GET /api/v1/auth/oidc/keys`
- **Token Revocation**: `POST /api/v1/auth/oidc/revoke`
- **Token Introspection**: `POST /api/v1/auth/oidc/introspect`
- **End Session**: `GET/POST /api/v1/auth/oidc/end_session`

#### Custom Endpoints

- **Login Page**: `GET /api/v1/auth/oidc/login`
- **Login Handler**: `POST /api/v1/auth/oidc/login`
- **Login Callback**: `GET /api/v1/auth/oidc/callback`
- **API Metadata**: `GET /api/metadata`

## Demo Users

The server comes with pre-configured demo users:

| Username | Password | Role |
|----------|----------|------|
| demo@example.com | password123 | user |
| admin@example.com | admin123 | admin |

## Demo Client

A default OIDC client is pre-configured:

- **Client ID**: `demo-client`
- **Client Secret**: `demo-secret`
- **Redirect URIs**:
  - `http://localhost:8080/callback`
  - `http://localhost:3000/callback`
- **Grant Types**: `authorization_code`, `refresh_token`
- **Response Types**: `code`
- **Scopes**: `openid`, `profile`, `email`

## Testing the OIDC Flow

### 1. Authorization Request

```bash
curl -X GET "http://localhost:8080/api/v1/auth/oidc/auth?response_type=code&client_id=demo-client&redirect_uri=http://localhost:8080/callback&scope=openid%20profile%20email&state=random-state"
```

This will redirect you to the login page.

### 2. Login

Visit the login page in your browser and use the demo credentials.

### 3. Token Exchange

After authorization, exchange the code for tokens:

```bash
curl -X POST http://localhost:8080/api/v1/auth/oidc/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=AUTH_CODE&redirect_uri=http://localhost:8080/callback&client_id=demo-client&client_secret=demo-secret"
```

### 4. UserInfo

Use the access token to get user information:

```bash
curl -X GET http://localhost:8080/api/v1/auth/oidc/userinfo \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

## Architecture

### Storage Layer

The implementation includes several storage components:

- **ClientStore** (`internal/storage/client.go`): Manages OIDC clients
- **AuthRequestStore** (`internal/storage/authrequest.go`): Handles authorization requests
- **UserInfoStore** (`internal/storage/userinfo.go`): Manages user information
- **OIDCStorage** (`internal/storage/oidcstorage.go`): Combined storage implementing all OIDC interfaces

### Router Integration

The OIDC server integrates with the existing router framework:

- **OIDCRouter** (`api/v1/oidc.go`): Implements the `router.ApiObject` interface
- Follows the same patterns as other API endpoints (cart, item, user)
- Provides versioned API paths (`/api/v1/auth/oidc/...`)

### Key Features

1. **Standards Compliant**: Full OIDC discovery and standard endpoints
2. **Secure**: RS256 JWT signing, PKCE support, proper token validation
3. **Extensible**: Easy to add new clients, users, and scopes
4. **Integrated**: Works with the existing demo-shop architecture
5. **Production Ready**: Proper error handling, logging, and security headers

## Configuration

The OIDC server can be configured via the `OIDCConfig` struct:

```go
config := &v1.OIDCConfig{
    Issuer:        "http://localhost:8080",
    Port:          8080,
    AllowInsecure: true, // Set to false in production
}
```

## Security Considerations

**Note**: This implementation is for demonstration purposes. For production use:

1. Use a proper database instead of in-memory storage
2. Implement proper key management and rotation
3. Use secure random keys for cryptography
4. Implement rate limiting and request validation
5. Use HTTPS in production (set `AllowInsecure: false`)
6. Implement proper session management
7. Add comprehensive logging and monitoring

## Development

### Adding New Clients

```go
client := &storage.Client{
    ClientID:                "your-client-id",
    ClientSecret:           "your-client-secret",
    ClientRedirectURIs:     []string{"http://your-app.com/callback"},
    ClientApplicationType:  op.ApplicationTypeWeb,
    ClientAuthMethod:       oidc.AuthMethodBasic,
    // ... other configuration
}

storage.CreateClient(ctx, client)
```

### Adding New Users

```go
user := &storage.OIDCUser{
    ID:                "user-id",
    Username:          "user@example.com",
    Password:          "hashed-password",
    Email:             "user@example.com",
    EmailVerified:     true,
    // ... other user data
}

userStore.CreateUser(ctx, user)
```

## Integration with Other Services

This OIDC server can be used to authenticate users for:

- Cart service API calls
- Item management operations
- User profile access
- Any other service in the demo-shop ecosystem

Simply validate the JWT tokens issued by this server in your other services.

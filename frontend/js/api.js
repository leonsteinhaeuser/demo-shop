// API Configuration - Using configurable URLs from environment variables
function getAPIConfig() {
    if (typeof window !== 'undefined' && window.API_CONFIG) {
        return window.API_CONFIG;
    }

    // Fallback configuration for local development
    return {
        ITEMS_SERVICE_URL: 'http://localhost:8081',
        USERS_SERVICE_URL: 'http://localhost:8084',
        CARTS_SERVICE_URL: 'http://localhost:8082',
        CHECKOUTS_SERVICE_URL: 'http://localhost:8085',
        CART_PRESENTATION_SERVICE_URL: 'http://localhost:8083'
    };
}

// API endpoint paths
const API_ENDPOINTS = {
    items: '/api/v1/core/items',
    users: '/api/v1/core/users',
    carts: '/api/v1/core/carts',
    checkouts: '/api/v1/core/checkouts',
    cartPresentation: '/api/v1/presentation/cart',
    auth: '/api/v1/auth'
};

// Build full URLs for each service
const API_SERVICES = {
    get items() {
        const config = getAPIConfig();
        return config.ITEMS_SERVICE_URL + API_ENDPOINTS.items;
    },
    get users() {
        const config = getAPIConfig();
        return config.USERS_SERVICE_URL + API_ENDPOINTS.users;
    },
    get carts() {
        const config = getAPIConfig();
        return config.CARTS_SERVICE_URL + API_ENDPOINTS.carts;
    },
    get checkouts() {
        const config = getAPIConfig();
        return config.CHECKOUTS_SERVICE_URL + API_ENDPOINTS.checkouts;
    },
    get cartPresentation() {
        const config = getAPIConfig();
        return config.CART_PRESENTATION_SERVICE_URL + API_ENDPOINTS.cartPresentation;
    },
    get auth() {
        const config = getAPIConfig();
        // Use any service URL for auth since they all point to the gateway
        return config.ITEMS_SERVICE_URL + API_ENDPOINTS.auth;
    }
};

class ApiClient {
    async request(url, options = {}) {
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers,
            },
            credentials: 'include', // Include cookies for session management
            ...options,
        };

        if (config.body && typeof config.body === 'object') {
            config.body = JSON.stringify(config.body);
        }

        try {
            const response = await fetch(url, config);

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                return await response.json();
            }

            return await response.text();
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    }

    // Authentication API
    async login(username, password) {
        return this.request(`${API_SERVICES.auth}/login`, {
            method: 'POST',
            body: { username, password },
        });
    }

    async logout() {
        return this.request(`${API_SERVICES.auth}/logout`, {
            method: 'POST',
        });
    }

    // Items API
    async getItems(page = 1, limit = 100) {
        return this.request(`${API_SERVICES.items}?page=${page}&limit=${limit}`);
    }

    async getItem(id) {
        return this.request(`${API_SERVICES.items}/${id}`);
    }

    async createItem(item) {
        return this.request(API_SERVICES.items, {
            method: 'POST',
            body: item,
        });
    }

    async updateItem(id, item) {
        return this.request(`${API_SERVICES.items}/${id}`, {
            method: 'PUT',
            body: item,
        });
    }

    async deleteItem(id) {
        const deleteItem = { id: id };
        return this.request(`${API_SERVICES.items}/${id}`, {
            method: 'DELETE',
            body: deleteItem,
        });
    }

    // Users API
    async getUsers(page = 1, limit = 100) {
        return this.request(`${API_SERVICES.users}?page=${page}&limit=${limit}`);
    }

    async getUser(id) {
        return this.request(`${API_SERVICES.users}/${id}`);
    }

    async createUser(user) {
        return this.request(API_SERVICES.users, {
            method: 'POST',
            body: user,
        });
    }

    async updateUser(id, user) {
        return this.request(`${API_SERVICES.users}/${id}`, {
            method: 'PUT',
            body: user,
        });
    }

    async deleteUser(id) {
        const deleteUser = { id: id };
        return this.request(`${API_SERVICES.users}/${id}`, {
            method: 'DELETE',
            body: deleteUser,
        });
    }

    // Carts API
    async getCarts() {
        // Note: The API doesn't have a list endpoint, so we'll return empty array
        // In a real implementation, you might need to add a list endpoint
        return [];
    }

    async getCart(id) {
        return this.request(`${API_SERVICES.carts}/${id}`);
    }

    async createCart(cart) {
        return this.request(API_SERVICES.carts, {
            method: 'POST',
            body: cart,
        });
    }

    async updateCart(id, cart) {
        return this.request(`${API_SERVICES.carts}/${id}`, {
            method: 'PUT',
            body: cart,
        });
    }

    async deleteCart(id) {
        const deleteCart = { id: id };
        return this.request(`${API_SERVICES.carts}/${id}`, {
            method: 'DELETE',
            body: deleteCart,
        });
    }

    // Cart Presentation API
    async getCartPresentation(id) {
        return this.request(`${API_SERVICES.cartPresentation}/${id}`);
    }

    // Checkouts API
    async getCheckouts() {
        // Note: The API doesn't have a list endpoint, so we'll return empty array
        // In a real implementation, you might need to add a list endpoint
        return [];
    }

    async getCheckout(id) {
        return this.request(`${API_SERVICES.checkouts}/${id}`);
    }

    async createCheckout(checkout) {
        return this.request(API_SERVICES.checkouts, {
            method: 'POST',
            body: checkout,
        });
    }

    async updateCheckout(id, checkout) {
        return this.request(`${API_SERVICES.checkouts}/${id}`, {
            method: 'PUT',
            body: checkout,
        });
    }

    async deleteCheckout(id) {
        const deleteCheckout = { id: id };
        return this.request(`${API_SERVICES.checkouts}/${id}`, {
            method: 'DELETE',
            body: deleteCheckout,
        });
    }

    // API Metadata
    async getApiMetadata() {
        const config = getAPIConfig();
        return this.request(`${config.ITEMS_SERVICE_URL}/api/metadata`);
    }
}

// Create global API client instance
window.apiClient = new ApiClient();

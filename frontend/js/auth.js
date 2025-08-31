// Authentication Management
class Auth {
    constructor() {
        this.currentUser = null;
        this.isAuthenticated = false;
        // Call async init in a non-blocking way
        this.init();
    }

    init() {
        // Check for stored session
        const storedUser = localStorage.getItem('demo-shop-user');
        const storedCartId = localStorage.getItem('demo-shop-cart-id');
        if (storedUser) {
            this.currentUser = JSON.parse(storedUser);
            this.isAuthenticated = true;
            this.updateUI();

            // Initialize cart for restored user session asynchronously
            if (window.cart && storedCartId) {
                window.cart.cartId = storedCartId;
                // Use setTimeout to ensure cart initialization happens after all synchronous initialization
                setTimeout(async () => {
                    try {
                        await window.cart.initializeCart();
                    } catch (error) {
                        console.error('Error initializing cart on session restore:', error);
                    }
                }, 100);
            }
        } else {
            // Ensure UI is updated for unauthenticated state
            this.updateUI();
        }

        this.setupEventListeners();
    }

    setupEventListeners() {
        // Login button
        document.getElementById('login-btn').addEventListener('click', () => {
            this.showLoginModal();
        });

        // Logout button
        document.getElementById('logout-btn').addEventListener('click', () => {
            this.logout();
        });

        // Login form
        document.getElementById('login-form').addEventListener('submit', (e) => {
            this.handleLogin(e);
        });

        // Register form
        document.getElementById('register-form').addEventListener('submit', (e) => {
            this.handleRegister(e);
        });

        // Register link
        document.getElementById('register-link').addEventListener('click', (e) => {
            e.preventDefault();
            this.showRegisterModal();
        });

        // Modal close events
        document.querySelectorAll('.close, .close-modal').forEach(button => {
            button.addEventListener('click', () => {
                this.closeModals();
            });
        });

        // Close modals when clicking outside
        window.addEventListener('click', (e) => {
            if (e.target.classList.contains('modal')) {
                this.closeModals();
            }
        });
    }

    async handleLogin(e) {
        e.preventDefault();
        const formData = new FormData(e.target);
        const credentials = {
            username: formData.get('username') || formData.get('email'), // Support both username and email fields for compatibility
            password: formData.get('password')
        };

        try {
            // Use the gateway authentication endpoint
            const response = await apiClient.login(credentials.username, credentials.password);
            
            this.currentUser = response.user;
            this.isAuthenticated = true;
            localStorage.setItem('demo-shop-user', JSON.stringify(response.user));
            localStorage.setItem('demo-shop-cart-id', response.cart_id);

            this.updateUI();
            this.closeModals();

            if (window.shop) {
                window.shop.showToast('Login successful!', 'success');
            }

            // Initialize user's cart with the cart ID from the login response
            if (window.cart) {
                window.cart.cartId = response.cart_id;
                await window.cart.initializeCart();
            }
            
            // Re-render products to show/hide Add to Cart buttons
            if (window.shop) {
                window.shop.renderProducts();
            }
            
            // Update navigation highlighting
            if (window.router) {
                window.router.highlightNav();
            }
        } catch (error) {
            if (window.shop) {
                window.shop.showToast('Invalid credentials', 'error');
            }
        }
    }

    async handleRegister(e) {
        e.preventDefault();
        const formData = new FormData(e.target);
        const userData = {
            username: formData.get('username'),
            email: formData.get('email'),
            password: formData.get('password'),
            preferred_name: formData.get('preferred_name') || null
        };

        try {
            const newUser = await apiClient.createUser(userData);

            if (window.shop) {
                window.shop.showToast('Registration successful! You can now login.', 'success');
            }

            this.showLoginModal();
        } catch (error) {
            if (window.shop) {
                window.shop.showToast('Registration failed: ' + error.message, 'error');
            }
        }
    }

    async logout() {
        console.log('Logout started, cart items before logout:', window.cart?.items?.length || 0);

        try {
            // Call the gateway logout endpoint
            await apiClient.logout();
        } catch (error) {
            console.error('Logout API call failed:', error);
            // Continue with local logout even if API call fails
        }

        this.currentUser = null;
        this.isAuthenticated = false;
        localStorage.removeItem('demo-shop-user');
        localStorage.removeItem('demo-shop-cart-id');

        // Clear cart UI and data
        if (window.cart) {
            window.cart.items = [];
            window.cart.cartId = null;
            window.cart.updateCartCount();
        }

        console.log('Logout completed, local cart cleared');

        this.updateUI();

        if (window.shop) {
            window.shop.showToast('Logged out successfully', 'info');
        }

        // Redirect to home if on admin page
        if (window.router.currentPath === '/shop/items') {
            window.router.navigate('/');
        }
        
        // Re-render products to show/hide Add to Cart buttons
        if (window.shop) {
            window.shop.renderProducts();
        }
        
        // Update navigation highlighting
        if (window.router) {
            window.router.highlightNav();
        }
    }

    updateUI() {
        const userName = document.getElementById('user-name');
        const loginBtn = document.getElementById('login-btn');
        const logoutBtn = document.getElementById('logout-btn');
        
        // Get containers for dynamic navigation
        const dynamicNavLinks = document.getElementById('dynamic-nav-links');
        const adminNavLink = document.getElementById('admin-nav-link');

        if (this.isAuthenticated && this.currentUser) {
            userName.textContent = this.currentUser.preferred_name || this.currentUser.username || 'User';
            loginBtn.classList.add('hidden');
            logoutBtn.classList.remove('hidden');

            // Render authenticated user navigation
            this.renderAuthenticatedNavigation(dynamicNavLinks, adminNavLink);
        } else {
            userName.textContent = 'Guest';
            loginBtn.classList.remove('hidden');
            logoutBtn.classList.add('hidden');
            
            // Clear navigation for unauthenticated users
            this.renderUnauthenticatedNavigation(dynamicNavLinks, adminNavLink);
        }
    }

    renderAuthenticatedNavigation(dynamicContainer, adminContainer) {
        // Render cart and profile navigation
        dynamicContainer.innerHTML = `
            <li><a href="#" onclick="router.navigate('/cart')" class="nav-link">
                    <i class="fas fa-shopping-cart"></i> Cart
                    <span id="cart-count" class="cart-count">0</span>
                </a></li>
            <li><a href="#" onclick="router.navigate('/profile')" class="nav-link">
                    <i class="fas fa-user"></i> Profile
                </a></li>
        `;

        // Render admin navigation if user is admin
        if (this.isAdmin()) {
            adminContainer.innerHTML = `
                <li>
                    <a href="#" onclick="router.navigate('/shop/items')" class="nav-link">
                        <i class="fas fa-boxes"></i> Manage Items
                    </a>
                </li>
            `;
        } else {
            adminContainer.innerHTML = '';
        }

        // Update cart count if cart exists
        if (window.cart) {
            // Use setTimeout to ensure DOM is updated before trying to update cart count
            setTimeout(() => {
                window.cart.updateCartCount();
            }, 0);
        }
    }

    renderUnauthenticatedNavigation(dynamicContainer, adminContainer) {
        // Clear all dynamic navigation
        dynamicContainer.innerHTML = '';
        adminContainer.innerHTML = '';
    }

    isAdmin() {
        return this.isAuthenticated &&
            this.currentUser &&
            this.currentUser.is_admin === true;
    }

    getCurrentUser() {
        return this.currentUser;
    }

    showLoginModal() {
        document.getElementById('login-modal').style.display = 'block';
        document.getElementById('register-modal').style.display = 'none';
    }

    showRegisterModal() {
        document.getElementById('register-modal').style.display = 'block';
        document.getElementById('login-modal').style.display = 'none';
    }

    closeModals() {
        document.querySelectorAll('.modal').forEach(modal => {
            modal.style.display = 'none';
        });
    }
}

// Initialize auth
window.auth = new Auth();

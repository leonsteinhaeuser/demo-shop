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
        if (storedUser) {
            this.currentUser = JSON.parse(storedUser);
            this.isAuthenticated = true;
            this.updateUI();
            
            // Initialize cart for restored user session asynchronously
            if (window.cart) {
                // Use setTimeout to ensure cart initialization happens after all synchronous initialization
                setTimeout(async () => {
                    try {
                        await window.cart.initializeCart();
                    } catch (error) {
                        console.error('Error initializing cart on session restore:', error);
                    }
                }, 100);
            }
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
            username: formData.get('username'),
            password: formData.get('password')
        };

        try {
            // For demo purposes, we'll simulate login
            // In a real app, you'd validate against the user service
            const users = await apiClient.getUsers();
            const user = users.find(u => u.username === credentials.username);
            
            if (user) {
                // Simulate password validation (in real app, this would be server-side)
                this.currentUser = user;
                this.isAuthenticated = true;
                localStorage.setItem('demo-shop-user', JSON.stringify(user));
                
                this.updateUI();
                this.closeModals();
                
                if (window.shop) {
                    window.shop.showToast('Login successful!', 'success');
                }
                
                // Initialize user's cart
                if (window.cart) {
                    await window.cart.initializeCart();
                }
            } else {
                if (window.shop) {
                    window.shop.showToast('Invalid username or password', 'error');
                }
            }
        } catch (error) {
            if (window.shop) {
                window.shop.showToast('Login failed: ' + error.message, 'error');
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

    logout() {
        console.log('Logout started, cart items before logout:', window.cart?.items?.length || 0);
        
        this.currentUser = null;
        this.isAuthenticated = false;
        localStorage.removeItem('demo-shop-user');
        
        // Clear cart UI only - don't clear cart data as it should persist in API
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
    }

    updateUI() {
        const userName = document.getElementById('user-name');
        const loginBtn = document.getElementById('login-btn');
        const logoutBtn = document.getElementById('logout-btn');
        const adminLink = document.getElementById('admin-link');

        if (this.isAuthenticated && this.currentUser) {
            userName.textContent = this.currentUser.preferred_name || this.currentUser.username || 'User';
            loginBtn.classList.add('hidden');
            logoutBtn.classList.remove('hidden');
            
            // Show admin link if user has is_admin = true
            if (this.isAdmin()) {
                adminLink.classList.remove('hidden');
                adminLink.classList.add('show');
            } else {
                adminLink.classList.add('hidden');
                adminLink.classList.remove('show');
            }
        } else {
            userName.textContent = 'Guest';
            loginBtn.classList.remove('hidden');
            logoutBtn.classList.add('hidden');
            adminLink.classList.add('hidden');
            adminLink.classList.remove('show');
        }
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

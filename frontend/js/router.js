// Simple Client-Side Router
class Router {
    constructor() {
        this.routes = {};
        this.currentPath = '/';
        this.init();
    }

    init() {
        // Handle browser back/forward buttons
        window.addEventListener('popstate', () => {
            this.navigate(window.location.pathname, false);
        });

        // Handle initial load
        this.navigate(window.location.pathname || '/', false);
    }

    addRoute(path, handler) {
        this.routes[path] = handler;
    }

    async navigate(path, pushState = true) {
        this.currentPath = path;

        if (pushState) {
            window.history.pushState({}, '', path);
        }

        // Hide all pages
        document.querySelectorAll('.page').forEach(page => {
            page.classList.remove('active');
        });

        // Update navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
        });

        // Show the appropriate page and handle routing
        switch (path) {
            case '/':
                this.showPage('shop-page');
                this.highlightNav();
                if (window.shop) {
                    window.shop.loadProducts();
                }
                break;
            case '/cart':
                if (window.auth && window.auth.isAuthenticated) {
                    this.showPage('cart-page');
                    this.highlightNav();
                    if (window.cart) {
                        await window.cart.loadCart();
                    }
                } else {
                    this.navigate('/');
                    if (window.auth) {
                        window.auth.showLoginModal();
                    }
                    if (window.shop) {
                        window.shop.showToast('Please login to view your cart', 'warning');
                    }
                }
                break;
            case '/profile':
                if (window.auth && window.auth.isAuthenticated) {
                    this.showPage('profile-page');
                    this.highlightNav();
                    if (window.shop) {
                        window.shop.loadProfile();
                    }
                } else {
                    this.navigate('/');
                    if (window.auth) {
                        window.auth.showLoginModal();
                    }
                    if (window.shop) {
                        window.shop.showToast('Please login to view your profile', 'warning');
                    }
                }
                break;
            case '/shop/items':
                if (window.auth && window.auth.isAdmin()) {
                    this.showPage('admin-items-page');
                    this.highlightNav();
                    if (window.shop) {
                        window.shop.loadAdminItems();
                    }
                } else {
                    this.navigate('/');
                    if (window.shop) {
                        window.shop.showToast('Access denied. Admin privileges required.', 'error');
                    }
                }
                break;
            default:
                this.navigate('/');
                break;
        }
    }

    showPage(pageId) {
        const page = document.getElementById(pageId);
        if (page) {
            page.classList.add('active');
        }
    }

    highlightNav(index) {
        // Clear all active states
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
        });
        
        // Set active state based on current path
        const navLinks = document.querySelectorAll('.nav-link');
        switch (this.currentPath) {
            case '/':
                if (navLinks[0]) navLinks[0].classList.add('active'); // Shop link
                break;
            case '/cart':
                const cartLink = document.querySelector('a[onclick*="/cart"]');
                if (cartLink) cartLink.classList.add('active');
                break;
            case '/profile':
                const profileLink = document.querySelector('a[onclick*="/profile"]');
                if (profileLink) profileLink.classList.add('active');
                break;
            case '/shop/items':
                const adminLink = document.querySelector('a[onclick*="/shop/items"]');
                if (adminLink) adminLink.classList.add('active');
                break;
        }
    }
}

// Initialize router
window.router = new Router();

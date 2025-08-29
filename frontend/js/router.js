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
                this.highlightNav(0);
                if (window.shop) {
                    window.shop.loadProducts();
                }
                break;
            case '/cart':
                this.showPage('cart-page');
                this.highlightNav(1);
                if (window.cart) {
                    await window.cart.loadCart();
                }
                break;
            case '/profile':
                this.showPage('profile-page');
                this.highlightNav(2);
                if (window.shop) {
                    window.shop.loadProfile();
                }
                break;
            case '/shop/items':
                if (window.auth && window.auth.isAdmin()) {
                    this.showPage('admin-items-page');
                    this.highlightNav(3);
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
        const navLinks = document.querySelectorAll('.nav-link');
        if (navLinks[index]) {
            navLinks[index].classList.add('active');
        }
    }
}

// Initialize router
window.router = new Router();

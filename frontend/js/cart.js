// Cart Management
class Cart {
    constructor() {
        this.items = [];
        this.cartId = null;
        this.taxRate = 0.08; // 8% tax
        this.init();
    }

    init() {
        this.updateCartCount();
        this.setupEventListeners();

        // Initialize cart for authenticated users
        setTimeout(() => {
            if (window.auth && window.auth.isAuthenticated) {
                this.initializeCart();
            }
        }, 100); // Small delay to ensure auth is ready
    }

    setupEventListeners() {
        // Checkout button
        const checkoutBtn = document.getElementById('checkout-btn');
        if (checkoutBtn) {
            checkoutBtn.addEventListener('click', () => {
                this.showCheckoutModal();
            });
        }

        // Confirm checkout button
        const confirmCheckoutBtn = document.getElementById('confirm-checkout-btn');
        if (confirmCheckoutBtn) {
            confirmCheckoutBtn.addEventListener('click', () => {
                this.processCheckout();
            });
        }
    }

    async initializeCart() {
        console.log('Cart initialization started, authenticated:', window.auth.isAuthenticated);

        if (!window.auth.isAuthenticated) {
            this.items = [];
            this.cartId = null;
            this.updateCartCount();
            console.log('User not authenticated, cart cleared');
            return;
        }

        try {
            // Get cart ID from localStorage (set during login)
            const storedCartId = localStorage.getItem('demo-shop-cart-id');

            if (storedCartId) {
                this.cartId = storedCartId;
                console.log('Using cart ID from login:', this.cartId);

                // Load the cart using the cart_id from login response
                await this.loadCartFromAPI(this.cartId);
                console.log('Cart loaded successfully, items:', this.items.length);
            } else {
                console.log('No cart ID found, creating new cart');
                await this.createNewCart();
            }
        } catch (error) {
            console.error('Error initializing cart:', error);
            // Fallback to creating a new cart
            await this.createNewCart();
        }

        this.updateCartCount();
    }

    async loadCartFromAPI(cartId) {
        try {
            console.log('Loading cart presentation from API with ID:', cartId);

            // Use cart presentation service to get fully populated cart with items and total
            const cartPresentation = await apiClient.getCartPresentation(cartId);
            console.log('Cart presentation loaded from API:', cartPresentation);

            if (cartPresentation && cartPresentation.items && cartPresentation.items.length > 0) {
                this.cartId = cartId;
                // Convert cart presentation items to our internal format
                this.items = cartPresentation.items.map(cartItem => ({
                    ...cartItem.item,
                    quantity: cartItem.quantity
                }));
                // Store the total price from the service (we can use this for validation)
                this.serverTotalPrice = cartPresentation.total_price;
                console.log('Cart items loaded from presentation service:', {
                    itemCount: this.items.length,
                    serverTotal: this.serverTotalPrice,
                    items: this.items.map(item => ({ id: item.id, name: item.name, quantity: item.quantity, price: item.price }))
                });
            } else if (cartPresentation && cartPresentation.items && cartPresentation.items.length === 0) {
                // Cart exists but is empty
                this.cartId = cartId;
                this.items = [];
                this.serverTotalPrice = 0;
                console.log('Cart found but is empty');
            } else {
                console.log('No existing cart found, creating new cart');
                await this.createNewCart();
            }
        } catch (error) {
            console.error('Error loading cart presentation from API:', error);
            // If the cart doesn't exist (create a new one)
            const errorMsg = error.message.toLowerCase();
            if (errorMsg.includes('404') || errorMsg.includes('not found') || errorMsg.includes('cart not found')) {
                console.log('Cart not found, creating new cart');
                await this.createNewCart();
            } else {
                console.log('Using fallback storage due to API error');
                this.loadCartFromStorage();
            }
        }
    }

    async createNewCart() {
        if (!window.auth.isAuthenticated) {
            return;
        }

        try {
            const user = window.auth.getCurrentUser();
            if (!user || !user.id) {
                console.error('No valid user found for cart creation');
                return;
            }

            // Use the cart ID from localStorage if available, otherwise generate one
            const storedCartId = localStorage.getItem('demo-shop-cart-id');
            const cartId = storedCartId || user.id; // Fallback to user.id if no cart_id stored

            const cartData = {
                id: cartId,
                owner_id: user.id,
                items: []
            };

            console.log('Creating new cart:', cartData);
            const createdCart = await apiClient.createCart(cartData);
            console.log('Cart created successfully:', createdCart);

            this.cartId = createdCart.id;
            this.items = [];
            
            // Store the cart ID for future use
            localStorage.setItem('demo-shop-cart-id', this.cartId);
        } catch (error) {
            console.error('Error creating new cart:', error);
            // Fallback to local storage
            const storedCartId = localStorage.getItem('demo-shop-cart-id');
            if (storedCartId) {
                this.cartId = storedCartId;
                this.items = [];
                console.log('Using stored cart ID as fallback:', this.cartId);
            }
        }
    }

    async addItem(product, quantity = 1) {
        if (!window.auth.isAuthenticated) {
            window.auth.showLoginModal();
            return;
        }

        const existingItemIndex = this.items.findIndex(item => item.id === product.id);

        if (existingItemIndex >= 0) {
            this.items[existingItemIndex].quantity += quantity;
        } else {
            this.items.push({
                ...product,
                quantity: quantity
            });
        }

        await this.syncCartWithAPI();
        this.updateCartCount();

        if (window.shop) {
            window.shop.showToast(`${product.name} added to cart!`, 'success');
        }
    }

    async removeItem(productId) {
        this.items = this.items.filter(item => item.id !== productId);
        await this.syncCartWithAPI();
        this.updateCartCount();
        this.loadCart(); // Refresh cart display if on cart page
    }

    async updateQuantity(productId, newQuantity) {
        const item = this.items.find(item => item.id === productId);
        if (item) {
            if (newQuantity <= 0) {
                await this.removeItem(productId);
            } else {
                item.quantity = newQuantity;
                await this.syncCartWithAPI();
                this.updateCartCount();
                this.loadCart(); // Refresh cart display
            }
        }
    }

    async clearCart() {
        this.items = [];
        await this.syncCartWithAPI();
        this.updateCartCount();
        this.loadCart(); // Refresh cart display if on cart page
    }

    async syncCartWithAPI() {
        if (!window.auth.isAuthenticated || !this.cartId) {
            return;
        }

        try {
            const cartItems = this.items.map(item => ({
                item_id: item.id,
                quantity: item.quantity
            }));

            const cartData = {
                id: this.cartId,
                owner_id: window.auth.getCurrentUser().id,
                items: cartItems
            };

            await apiClient.updateCart(this.cartId, cartData);
            console.log('Cart synced with API:', cartData);
        } catch (error) {
            console.error('Error syncing cart with API:', error);
            // Fallback to localStorage for offline functionality
            this.saveCartToStorage();
        }
    }

    getSubtotal() {
        return this.items.reduce((total, item) => total + (item.price * item.quantity), 0);
    }

    getTax() {
        return this.getSubtotal() * this.taxRate;
    }

    getTotal() {
        const clientTotal = this.getSubtotal() + this.getTax();
        
        // Validate against server total if available
        if (this.serverTotalPrice !== undefined) {
            const serverSubtotal = this.serverTotalPrice;
            const serverTotal = serverSubtotal + (serverSubtotal * this.taxRate);
            
            // Log discrepancy if there's a significant difference (more than 1 cent)
            const difference = Math.abs(clientTotal - serverTotal);
            if (difference > 0.01) {
                console.warn('Cart total discrepancy detected:', {
                    clientTotal: clientTotal.toFixed(2),
                    serverTotal: serverTotal.toFixed(2),
                    difference: difference.toFixed(2)
                });
            }
        }
        
        return clientTotal;
    }

    getServerSubtotal() {
        return this.serverTotalPrice || this.getSubtotal();
    }

    updateCartCount() {
        const totalItems = this.items.reduce((total, item) => total + item.quantity, 0);
        const cartCountElement = document.getElementById('cart-count');
        if (cartCountElement) {
            cartCountElement.textContent = totalItems;
            cartCountElement.style.display = totalItems > 0 ? 'block' : 'none';
        }
    }

    async loadCart() {
        // First, refresh cart data from API if user is authenticated
        if (window.auth && window.auth.isAuthenticated) {
            try {
                // Use the stored cart ID instead of user ID
                const storedCartId = localStorage.getItem('demo-shop-cart-id');
                if (storedCartId) {
                    console.log('Refreshing cart from API with cart ID:', storedCartId);
                    await this.loadCartFromAPI(storedCartId);
                } else {
                    console.log('No cart ID found, initializing cart');
                    await this.initializeCart();
                }
            } catch (error) {
                console.error('Error refreshing cart from API:', error);
                // Fallback to localStorage
                this.loadCartFromStorage();
            }
        }

        // Now display the cart items
        const cartItemsContainer = document.getElementById('cart-items');
        if (!cartItemsContainer) return;

        if (this.items.length === 0) {
            cartItemsContainer.innerHTML = `
                <div class="text-center text-muted">
                    <i class="fas fa-shopping-cart" style="font-size: 3rem; margin-bottom: 1rem;"></i>
                    <h3>Your cart is empty</h3>
                    <p>Add some items to get started!</p>
                    <button onclick="router.navigate('/')" class="btn btn-primary">
                        <i class="fas fa-store"></i> Continue Shopping
                    </button>
                </div>
            `;
        } else {
            cartItemsContainer.innerHTML = this.items.map(item => `
                <div class="cart-item">
                    <div class="cart-item-image">
                        <i class="fas fa-box"></i>
                    </div>
                    <div class="cart-item-details">
                        <div class="cart-item-name">${item.name}</div>
                        <div class="cart-item-price">$${item.price.toFixed(2)} each</div>
                        <div class="cart-item-controls">
                            <div class="quantity-controls">
                                <button class="quantity-btn" onclick="cart.updateQuantity('${item.id}', ${item.quantity - 1})">
                                    <i class="fas fa-minus"></i>
                                </button>
                                <input type="number" class="quantity-input" value="${item.quantity}" 
                                       onchange="cart.updateQuantity('${item.id}', parseInt(this.value))" min="1">
                                <button class="quantity-btn" onclick="cart.updateQuantity('${item.id}', ${item.quantity + 1})">
                                    <i class="fas fa-plus"></i>
                                </button>
                            </div>
                            <button class="btn btn-danger btn-sm" onclick="cart.removeItem('${item.id}')">
                                <i class="fas fa-trash"></i> Remove
                            </button>
                        </div>
                    </div>
                    <div class="cart-item-total">
                        <strong>$${(item.price * item.quantity).toFixed(2)}</strong>
                    </div>
                </div>
            `).join('');
        }

        // Update summary
        this.updateSummary();
        this.updateCartCount(); // Update cart count after loading
    }

    updateSummary() {
        const subtotalElement = document.getElementById('cart-subtotal');
        const taxElement = document.getElementById('cart-tax');
        const totalElement = document.getElementById('cart-total');

        if (subtotalElement) {
            subtotalElement.textContent = `$${this.getSubtotal().toFixed(2)}`;
            
            // Add server subtotal for comparison in development
            if (this.serverTotalPrice !== undefined && window.location.hostname === 'localhost') {
                const serverSubtotal = this.serverTotalPrice;
                subtotalElement.title = `Server subtotal: $${serverSubtotal.toFixed(2)}`;
            }
        }
        if (taxElement) {
            taxElement.textContent = `$${this.getTax().toFixed(2)}`;
        }
        if (totalElement) {
            totalElement.textContent = `$${this.getTotal().toFixed(2)}`;
        }
    }

    showCheckoutModal() {
        if (!window.auth.isAuthenticated) {
            window.auth.showLoginModal();
            return;
        }

        if (this.items.length === 0) {
            if (window.shop) {
                window.shop.showToast('Your cart is empty!', 'warning');
            }
            return;
        }

        const modal = document.getElementById('checkout-modal');
        const itemsSummary = document.getElementById('checkout-items-summary');
        const totalAmount = document.getElementById('checkout-total-amount');

        itemsSummary.innerHTML = this.items.map(item => `
            <div style="display: flex; justify-content: space-between; margin-bottom: 0.5rem;">
                <span>${item.name} (${item.quantity}x)</span>
                <span>$${(item.price * item.quantity).toFixed(2)}</span>
            </div>
        `).join('');

        totalAmount.textContent = `$${this.getTotal().toFixed(2)}`;
        modal.style.display = 'block';
    }

    async processCheckout() {
        if (!window.auth.isAuthenticated) {
            if (window.shop) {
                window.shop.showToast('Please login to checkout', 'error');
            }
            return;
        }

        try {
            window.shop.showLoading();
            const user = window.auth.getCurrentUser();

            // Create checkout data
            const checkoutData = {
                user_id: user.id,
                cart_id: this.cartId,
                total: this.getTotal(),
                status: 'completed'
            };

            // Call the checkout service
            const checkout = await apiClient.createCheckout(checkoutData);
            console.log('Checkout created:', checkout);

            // Clear the cart after successful checkout
            await this.clearCart();

            // Close modal
            document.getElementById('checkout-modal').style.display = 'none';

            if (window.shop) {
                window.shop.showToast('Order placed successfully! Thank you for your purchase.', 'success');
            }

            // Redirect to shop
            window.router.navigate('/');

        } catch (error) {
            console.error('Checkout error:', error);
            if (window.shop) {
                window.shop.showToast('Checkout failed: ' + error.message, 'error');
            }
        } finally {
            window.shop.hideLoading();
        }
    }

    saveCartToStorage() {
        // Fallback for offline functionality
        if (window.auth.isAuthenticated && this.cartId) {
            localStorage.setItem(`cart-${this.cartId}`, JSON.stringify(this.items));
        }
    }

    loadCartFromStorage() {
        // Fallback for offline functionality
        if (window.auth.isAuthenticated && this.cartId) {
            const storedCart = localStorage.getItem(`cart-${this.cartId}`);
            if (storedCart) {
                this.items = JSON.parse(storedCart);
            }
        }
    }
}

// Initialize cart
window.cart = new Cart();

// Main Shop Application
class Shop {
    constructor() {
        this.products = [];
        this.filteredProducts = [];
        this.editingItemId = null;
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadProducts();
    }

    setupEventListeners() {
        // Search functionality
        const searchInput = document.getElementById('search-input');
        const searchBtn = document.getElementById('search-btn');
        const sortSelect = document.getElementById('sort-select');

        if (searchInput && searchBtn) {
            searchInput.addEventListener('input', () => this.filterProducts());
            searchBtn.addEventListener('click', () => this.filterProducts());
        }

        if (sortSelect) {
            sortSelect.addEventListener('change', () => this.sortProducts());
        }

        // Profile form
        const profileForm = document.getElementById('profile-form');
        if (profileForm) {
            profileForm.addEventListener('submit', (e) => this.handleProfileUpdate(e));
        }

        // Admin item form
        const itemForm = document.getElementById('item-form');
        if (itemForm) {
            itemForm.addEventListener('submit', (e) => this.handleItemSubmit(e));
        }

        // Add item button
        const addItemBtn = document.getElementById('add-item-btn');
        if (addItemBtn) {
            addItemBtn.addEventListener('click', () => this.openItemModal());
        }

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

    async loadProducts() {
        try {
            this.showLoading();
            this.products = await apiClient.getItems() || [];
            this.filteredProducts = [...this.products];
            this.renderProducts();
        } catch (error) {
            console.error('Error loading products:', error);
            this.showToast('Error loading products: ' + error.message, 'error');
        } finally {
            this.hideLoading();
        }
    }

    renderProducts() {
        const productsGrid = document.getElementById('products-grid');
        if (!productsGrid) return;

        if (this.filteredProducts.length === 0) {
            productsGrid.innerHTML = `
                <div class="text-center text-muted" style="grid-column: 1 / -1;">
                    <i class="fas fa-search" style="font-size: 3rem; margin-bottom: 1rem;"></i>
                    <h3>No products found</h3>
                    <p>Try adjusting your search criteria</p>
                </div>
            `;
            return;
        }

        productsGrid.innerHTML = this.filteredProducts.map(product => `
            <div class="product-card">
                <div class="product-image">
                    <i class="fas fa-box"></i>
                </div>
                <div class="product-info">
                    <div class="product-name">${product.name}</div>
                    <div class="product-description">${product.description}</div>
                    <div class="product-price">$${product.price.toFixed(2)}</div>
                    <div class="product-stock">
                        ${product.quantity > 0 ? `${product.quantity} in stock` : 'Out of stock'}
                    </div>
                    <div class="product-actions">
                        <button class="btn btn-primary" 
                                onclick="shop.addToCart('${product.id}')"
                                ${product.quantity <= 0 ? 'disabled' : ''}>
                            <i class="fas fa-cart-plus"></i> Add to Cart
                        </button>
                    </div>
                </div>
            </div>
        `).join('');
    }

    filterProducts() {
        const searchTerm = document.getElementById('search-input')?.value.toLowerCase() || '';
        
        this.filteredProducts = this.products.filter(product =>
            product.name.toLowerCase().includes(searchTerm) ||
            product.description.toLowerCase().includes(searchTerm)
        );

        this.sortProducts();
        this.renderProducts();
    }

    sortProducts() {
        const sortValue = document.getElementById('sort-select')?.value || 'name';

        this.filteredProducts.sort((a, b) => {
            switch (sortValue) {
                case 'price-low':
                    return a.price - b.price;
                case 'price-high':
                    return b.price - a.price;
                case 'name':
                default:
                    return a.name.localeCompare(b.name);
            }
        });

        this.renderProducts();
    }

    async addToCart(productId) {
        if (!window.auth.isAuthenticated) {
            window.auth.showLoginModal();
            return;
        }

        const product = this.products.find(p => p.id === productId);
        if (product && product.quantity > 0) {
            try {
                this.showLoading();
                await window.cart.addItem(product, 1);
            } catch (error) {
                console.error('Error adding to cart:', error);
                this.showToast('Failed to add item to cart: ' + error.message, 'error');
            } finally {
                this.hideLoading();
            }
        } else {
            this.showToast('Product is out of stock', 'warning');
        }
    }

    async loadProfile() {
        if (!window.auth.isAuthenticated) {
            window.router.navigate('/');
            return;
        }

        const user = window.auth.getCurrentUser();
        if (user) {
            document.getElementById('profile-username').value = user.username || '';
            document.getElementById('profile-email').value = user.email || '';
            document.getElementById('profile-preferred-name').value = user.preferred_name || '';
            document.getElementById('profile-given-name').value = user.given_name || '';
            document.getElementById('profile-family-name').value = user.family_name || '';
            document.getElementById('profile-password').value = '';
        }
    }

    async handleProfileUpdate(e) {
        e.preventDefault();
        
        if (!window.auth.isAuthenticated) {
            this.showToast('Please login to update profile', 'error');
            return;
        }

        const formData = new FormData(e.target);
        const user = window.auth.getCurrentUser();
        
        const updatedUser = {
            id: user.id,
            username: formData.get('username'),
            email: formData.get('email'),
            preferred_name: formData.get('preferred_name') || null,
            given_name: formData.get('given_name') || null,
            family_name: formData.get('family_name') || null
        };

        const password = formData.get('password');
        if (password) {
            updatedUser.password = password;
        }

        try {
            this.showLoading();
            await apiClient.updateUser(user.id, updatedUser);
            
            // Update local user data
            const newUserData = { ...user, ...updatedUser };
            window.auth.currentUser = newUserData;
            localStorage.setItem('demo-shop-user', JSON.stringify(newUserData));
            window.auth.updateUI();
            
            this.showToast('Profile updated successfully!', 'success');
        } catch (error) {
            console.error('Error updating profile:', error);
            this.showToast('Error updating profile: ' + error.message, 'error');
        } finally {
            this.hideLoading();
        }
    }

    async loadAdminItems() {
        if (!window.auth.isAdmin()) {
            window.router.navigate('/');
            return;
        }

        try {
            this.showLoading();
            this.products = await apiClient.getItems() || [];
            this.renderAdminItems();
        } catch (error) {
            console.error('Error loading admin items:', error);
            this.showToast('Error loading items: ' + error.message, 'error');
        } finally {
            this.hideLoading();
        }
    }

    renderAdminItems() {
        const tbody = document.querySelector('#admin-items-table tbody');
        if (!tbody) return;

        tbody.innerHTML = this.products.map(item => `
            <tr>
                <td>${item.name}</td>
                <td>${item.description}</td>
                <td>$${item.price.toFixed(2)}</td>
                <td>${item.quantity}</td>
                <td>${item.location}</td>
                <td>${new Date(item.created_at).toLocaleDateString()}</td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="shop.editItem('${item.id}')">
                        <i class="fas fa-edit"></i> Edit
                    </button>
                    <button class="btn btn-sm btn-danger" onclick="shop.deleteItem('${item.id}')">
                        <i class="fas fa-trash"></i> Delete
                    </button>
                </td>
            </tr>
        `).join('');
    }

    openItemModal(item = null) {
        this.editingItemId = item ? item.id : null;
        const modal = document.getElementById('item-modal');
        const title = document.getElementById('item-modal-title');
        const form = document.getElementById('item-form');

        title.textContent = item ? 'Edit Item' : 'Add Item';
        
        if (item) {
            document.getElementById('item-name').value = item.name;
            document.getElementById('item-description').value = item.description;
            document.getElementById('item-price').value = item.price;
            document.getElementById('item-quantity').value = item.quantity;
            document.getElementById('item-location').value = item.location;
        } else {
            form.reset();
        }

        modal.style.display = 'block';
    }

    async handleItemSubmit(e) {
        e.preventDefault();
        const formData = new FormData(e.target);
        const item = {
            name: formData.get('name'),
            description: formData.get('description'),
            price: parseFloat(formData.get('price')),
            quantity: parseInt(formData.get('quantity')),
            location: formData.get('location')
        };

        try {
            this.showLoading();
            
            if (this.editingItemId) {
                item.id = this.editingItemId;
                await apiClient.updateItem(this.editingItemId, item);
                this.showToast('Item updated successfully', 'success');
            } else {
                await apiClient.createItem(item);
                this.showToast('Item created successfully', 'success');
            }
            
            this.closeModals();
            await this.loadAdminItems();
            await this.loadProducts(); // Refresh shop products too
        } catch (error) {
            console.error('Error saving item:', error);
            this.showToast('Error saving item: ' + error.message, 'error');
        } finally {
            this.hideLoading();
        }
    }

    async editItem(id) {
        const item = this.products.find(i => i.id === id);
        if (item) {
            this.openItemModal(item);
        }
    }

    async deleteItem(id) {
        if (confirm('Are you sure you want to delete this item?')) {
            try {
                this.showLoading();
                await apiClient.deleteItem(id);
                this.showToast('Item deleted successfully', 'success');
                await this.loadAdminItems();
                await this.loadProducts(); // Refresh shop products too
            } catch (error) {
                console.error('Error deleting item:', error);
                this.showToast('Error deleting item: ' + error.message, 'error');
            } finally {
                this.hideLoading();
            }
        }
    }

    closeModals() {
        document.querySelectorAll('.modal').forEach(modal => {
            modal.style.display = 'none';
        });
        this.editingItemId = null;
    }

    showLoading() {
        document.getElementById('loading').classList.remove('hidden');
    }

    hideLoading() {
        document.getElementById('loading').classList.add('hidden');
    }

    showToast(message, type = 'info') {
        const container = document.getElementById('toast-container');
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;
        
        container.appendChild(toast);
        
        // Auto remove after 5 seconds
        setTimeout(() => {
            toast.remove();
        }, 5000);
    }
}

// Initialize shop when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.shop = new Shop();
    
    // Initialize cart after a short delay to ensure auth is ready
    setTimeout(() => {
        if (window.auth && window.auth.isAuthenticated) {
            window.cart.initializeCart();
        }
    }, 100);
});

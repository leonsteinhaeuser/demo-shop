// Dashboard Application
class Dashboard {
    constructor() {
        this.currentView = 'dashboard';
        this.data = {
            items: [],
            users: [],
            carts: [],
            checkouts: []
        };
        this.editingId = null;
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadDashboard();
    }

    setupEventListeners() {
        // Navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const view = e.currentTarget.dataset.view;
                this.switchView(view);
            });
        });

        // Refresh button
        document.getElementById('refresh-btn').addEventListener('click', () => {
            this.loadCurrentView();
        });

        // Add buttons
        document.getElementById('add-item-btn').addEventListener('click', () => {
            this.openItemModal();
        });

        document.getElementById('add-user-btn').addEventListener('click', () => {
            this.openUserModal();
        });

        document.getElementById('add-cart-btn').addEventListener('click', () => {
            this.openCartModal();
        });

        document.getElementById('add-checkout-btn').addEventListener('click', () => {
            this.openCheckoutModal();
        });

        // Modal close events
        document.querySelectorAll('.close, .close-modal').forEach(button => {
            button.addEventListener('click', () => {
                this.closeModals();
            });
        });

        // Form submissions
        document.getElementById('item-form').addEventListener('submit', (e) => {
            this.handleItemSubmit(e);
        });

        document.getElementById('user-form').addEventListener('submit', (e) => {
            this.handleUserSubmit(e);
        });

        // Close modals when clicking outside
        window.addEventListener('click', (e) => {
            if (e.target.classList.contains('modal')) {
                this.closeModals();
            }
        });
    }

    switchView(view) {
        // Update navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
        });
        document.querySelector(`[data-view="${view}"]`).classList.add('active');

        // Update views
        document.querySelectorAll('.view').forEach(viewElement => {
            viewElement.classList.remove('active');
        });
        document.getElementById(`${view}-view`).classList.add('active');

        // Update page title
        const titles = {
            dashboard: 'Dashboard',
            items: 'Items',
            users: 'Users',
            carts: 'Carts',
            checkouts: 'Checkouts'
        };
        document.getElementById('page-title').textContent = titles[view];

        this.currentView = view;
        this.loadCurrentView();
    }

    async loadCurrentView() {
        this.showLoading();
        try {
            switch (this.currentView) {
                case 'dashboard':
                    await this.loadDashboard();
                    break;
                case 'items':
                    await this.loadItems();
                    break;
                case 'users':
                    await this.loadUsers();
                    break;
                case 'carts':
                    await this.loadCarts();
                    break;
                case 'checkouts':
                    await this.loadCheckouts();
                    break;
            }
        } catch (error) {
            this.showToast('Error loading data: ' + error.message, 'error');
        } finally {
            this.hideLoading();
        }
    }

    async loadDashboard() {
        try {
            // Load all data for dashboard stats
            const [items, users] = await Promise.all([
                apiClient.getItems(),
                apiClient.getUsers()
            ]);

            this.data.items = items || [];
            this.data.users = users || [];

            // Update stats
            document.getElementById('total-items').textContent = this.data.items.length;
            document.getElementById('total-users').textContent = this.data.users.length;
            document.getElementById('total-carts').textContent = '0'; // API doesn't support listing
            document.getElementById('total-checkouts').textContent = '0'; // API doesn't support listing

            // Update recent activity
            this.updateRecentActivity();
        } catch (error) {
            console.error('Error loading dashboard:', error);
        }
    }

    updateRecentActivity() {
        const activityContainer = document.getElementById('recent-activity');
        const activities = [];

        // Add recent items
        this.data.items.slice(-5).forEach(item => {
            activities.push({
                title: `New item added: ${item.name}`,
                time: new Date(item.created_at).toLocaleString(),
                timestamp: new Date(item.created_at).getTime()
            });
        });

        // Add recent users
        this.data.users.slice(-5).forEach(user => {
            activities.push({
                title: `New user registered: ${user.username || 'Unknown'}`,
                time: new Date(user.created_at).toLocaleString(),
                timestamp: new Date(user.created_at).getTime()
            });
        });

        // Sort by timestamp (newest first)
        activities.sort((a, b) => b.timestamp - a.timestamp);

        // Display activities
        activityContainer.innerHTML = activities.slice(0, 10).map(activity => `
            <div class="activity-item">
                <div class="activity-title">${activity.title}</div>
                <div class="activity-time">${activity.time}</div>
            </div>
        `).join('');
    }

    async loadItems() {
        try {
            this.data.items = await apiClient.getItems() || [];
            this.renderItemsTable();
        } catch (error) {
            this.showToast('Error loading items: ' + error.message, 'error');
        }
    }

    renderItemsTable() {
        const tbody = document.querySelector('#items-table tbody');
        tbody.innerHTML = this.data.items.map(item => `
            <tr>
                <td>${item.name}</td>
                <td>${item.description}</td>
                <td>$${item.price.toFixed(2)}</td>
                <td>${item.quantity}</td>
                <td>${item.location}</td>
                <td>${new Date(item.created_at).toLocaleDateString()}</td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="dashboard.editItem('${item.id}')">
                        <i class="fas fa-edit"></i> Edit
                    </button>
                    <button class="btn btn-sm btn-danger" onclick="dashboard.deleteItem('${item.id}')">
                        <i class="fas fa-trash"></i> Delete
                    </button>
                </td>
            </tr>
        `).join('');
    }

    async loadUsers() {
        try {
            this.data.users = await apiClient.getUsers() || [];
            this.renderUsersTable();
        } catch (error) {
            this.showToast('Error loading users: ' + error.message, 'error');
        }
    }

    renderUsersTable() {
        const tbody = document.querySelector('#users-table tbody');
        tbody.innerHTML = this.data.users.map(user => `
            <tr>
                <td>${user.username || 'N/A'}</td>
                <td>${user.email || 'N/A'}</td>
                <td>
                    <span class="status-badge ${user.email_verified ? 'status-completed' : 'status-pending'}">
                        ${user.email_verified ? 'Verified' : 'Pending'}
                    </span>
                </td>
                <td>${user.preferred_name || 'N/A'}</td>
                <td>${new Date(user.created_at).toLocaleDateString()}</td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="dashboard.editUser('${user.id}')">
                        <i class="fas fa-edit"></i> Edit
                    </button>
                    <button class="btn btn-sm btn-danger" onclick="dashboard.deleteUser('${user.id}')">
                        <i class="fas fa-trash"></i> Delete
                    </button>
                </td>
            </tr>
        `).join('');
    }

    async loadCarts() {
        try {
            this.data.carts = await apiClient.getCarts() || [];
            this.renderCartsTable();
        } catch (error) {
            this.showToast('Error loading carts: ' + error.message, 'error');
        }
    }

    renderCartsTable() {
        const tbody = document.querySelector('#carts-table tbody');
        tbody.innerHTML = this.data.carts.map(cart => `
            <tr>
                <td>${cart.id}</td>
                <td>${cart.owner_id}</td>
                <td>${cart.items ? cart.items.length : 0}</td>
                <td>${new Date(cart.created_at).toLocaleDateString()}</td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="dashboard.viewCart('${cart.id}')">
                        <i class="fas fa-eye"></i> View
                    </button>
                    <button class="btn btn-sm btn-danger" onclick="dashboard.deleteCart('${cart.id}')">
                        <i class="fas fa-trash"></i> Delete
                    </button>
                </td>
            </tr>
        `).join('');
    }

    async loadCheckouts() {
        try {
            this.data.checkouts = await apiClient.getCheckouts() || [];
            this.renderCheckoutsTable();
        } catch (error) {
            this.showToast('Error loading checkouts: ' + error.message, 'error');
        }
    }

    renderCheckoutsTable() {
        const tbody = document.querySelector('#checkouts-table tbody');
        tbody.innerHTML = this.data.checkouts.map(checkout => `
            <tr>
                <td>${checkout.id}</td>
                <td>${checkout.user_id}</td>
                <td>${checkout.cart_id}</td>
                <td>$${checkout.total.toFixed(2)}</td>
                <td>
                    <span class="status-badge status-${checkout.status}">
                        ${checkout.status}
                    </span>
                </td>
                <td>${new Date(checkout.created_at).toLocaleDateString()}</td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="dashboard.editCheckout('${checkout.id}')">
                        <i class="fas fa-edit"></i> Edit
                    </button>
                    <button class="btn btn-sm btn-danger" onclick="dashboard.deleteCheckout('${checkout.id}')">
                        <i class="fas fa-trash"></i> Delete
                    </button>
                </td>
            </tr>
        `).join('');
    }

    // Modal functions
    openItemModal(item = null) {
        this.editingId = item ? item.id : null;
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

    openUserModal(user = null) {
        this.editingId = user ? user.id : null;
        const modal = document.getElementById('user-modal');
        const title = document.getElementById('user-modal-title');
        const form = document.getElementById('user-form');

        title.textContent = user ? 'Edit User' : 'Add User';
        
        if (user) {
            document.getElementById('user-username').value = user.username || '';
            document.getElementById('user-email').value = user.email || '';
            document.getElementById('user-preferred-name').value = user.preferred_name || '';
            document.getElementById('user-given-name').value = user.given_name || '';
            document.getElementById('user-family-name').value = user.family_name || '';
            // Don't pre-fill password for security
            document.getElementById('user-password').value = '';
        } else {
            form.reset();
        }

        modal.style.display = 'block';
    }

    openCartModal() {
        // TODO: Implement cart creation modal
        this.showToast('Cart creation not implemented in this demo', 'warning');
    }

    openCheckoutModal() {
        // TODO: Implement checkout creation modal
        this.showToast('Checkout creation not implemented in this demo', 'warning');
    }

    closeModals() {
        document.querySelectorAll('.modal').forEach(modal => {
            modal.style.display = 'none';
        });
        this.editingId = null;
    }

    // Form handlers
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
            if (this.editingId) {
                item.id = this.editingId;
                await apiClient.updateItem(this.editingId, item);
                this.showToast('Item updated successfully', 'success');
            } else {
                await apiClient.createItem(item);
                this.showToast('Item created successfully', 'success');
            }
            
            this.closeModals();
            this.loadItems();
        } catch (error) {
            this.showToast('Error saving item: ' + error.message, 'error');
        }
    }

    async handleUserSubmit(e) {
        e.preventDefault();
        const formData = new FormData(e.target);
        const user = {
            username: formData.get('username'),
            email: formData.get('email'),
            preferred_name: formData.get('preferred_name') || null,
            given_name: formData.get('given_name') || null,
            family_name: formData.get('family_name') || null,
        };

        const password = formData.get('password');
        if (password) {
            user.password = password;
        }

        try {
            if (this.editingId) {
                user.id = this.editingId;
                await apiClient.updateUser(this.editingId, user);
                this.showToast('User updated successfully', 'success');
            } else {
                await apiClient.createUser(user);
                this.showToast('User created successfully', 'success');
            }
            
            this.closeModals();
            this.loadUsers();
        } catch (error) {
            this.showToast('Error saving user: ' + error.message, 'error');
        }
    }

    // Action handlers
    async editItem(id) {
        const item = this.data.items.find(i => i.id === id);
        if (item) {
            this.openItemModal(item);
        }
    }

    async deleteItem(id) {
        if (confirm('Are you sure you want to delete this item?')) {
            try {
                await apiClient.deleteItem(id);
                this.showToast('Item deleted successfully', 'success');
                this.loadItems();
            } catch (error) {
                this.showToast('Error deleting item: ' + error.message, 'error');
            }
        }
    }

    async editUser(id) {
        const user = this.data.users.find(u => u.id === id);
        if (user) {
            this.openUserModal(user);
        }
    }

    async deleteUser(id) {
        if (confirm('Are you sure you want to delete this user?')) {
            try {
                await apiClient.deleteUser(id);
                this.showToast('User deleted successfully', 'success');
                this.loadUsers();
            } catch (error) {
                this.showToast('Error deleting user: ' + error.message, 'error');
            }
        }
    }

    async viewCart(id) {
        try {
            const cart = await apiClient.getCart(id);
            const presentation = await apiClient.getCartPresentation(id);
            
            // TODO: Show cart details in a modal
            console.log('Cart:', cart);
            console.log('Cart Presentation:', presentation);
            this.showToast('Cart details logged to console', 'info');
        } catch (error) {
            this.showToast('Error loading cart: ' + error.message, 'error');
        }
    }

    async deleteCart(id) {
        if (confirm('Are you sure you want to delete this cart?')) {
            try {
                await apiClient.deleteCart(id);
                this.showToast('Cart deleted successfully', 'success');
                this.loadCarts();
            } catch (error) {
                this.showToast('Error deleting cart: ' + error.message, 'error');
            }
        }
    }

    // Utility functions
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

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.dashboard = new Dashboard();
});

// 当前页码和每页显示数量
let currentPage = 1;
let pageSize = 20;
let totalPages = 1;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    // 加载用户列表
    loadUsers();

    // 初始化事件监听
    initEventListeners();
});

// 初始化事件监听器
function initEventListeners() {
    // 添加用户按钮
    document.getElementById('addUserBtn').addEventListener('click', () => {
        showUserModal();
    });

    // 分页大小选择
    document.getElementById('userPageSize').addEventListener('change', (e) => {
        pageSize = parseInt(e.target.value);
        currentPage = 1;
        loadUsers();
    });

    // 分页按钮
    document.getElementById('userPrevPage').addEventListener('click', () => {
        if (currentPage > 1) {
            currentPage--;
            loadUsers();
        }
    });

    document.getElementById('userNextPage').addEventListener('click', () => {
        if (currentPage < totalPages) {
            currentPage++;
            loadUsers();
        }
    });

    // 保存用户按钮
    document.getElementById('saveUserBtn').addEventListener('click', saveUser);

    // 关闭模态框按钮
    document.querySelectorAll('.close-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            document.getElementById('userFormModal').style.display = 'none';
        });
    });
}

// 加载用户列表
async function loadUsers() {
    try {
        const response = await fetch(`/api/users?page=${currentPage}&pageSize=${pageSize}`);
        const data = await response.json();

        if (data.error) {
            showMessage(data.error, 'error');
            return;
        }

        totalPages = Math.ceil(data.total / pageSize);
        renderUsers(data.users);
        updatePagination();
    } catch (error) {
        showMessage('加载用户列表失败', 'error');
    }
}

// 渲染用户列表
function renderUsers(users) {
    const tbody = document.getElementById('usersTableBody');
    tbody.innerHTML = '';

    users.forEach(user => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${user.id}</td>
            <td>${user.username}</td>
            <td>${user.name}</td>
            <td>${user.email}</td>
            <td>${user.role === 'admin' ? '管理员' : '普通用户'}</td>
            <td>
                <span class="user-status ${user.is_active ? 'status-active' : 'status-inactive'}">
                    ${user.is_active ? '启用' : '禁用'}
                </span>
            </td>
            <td>${formatDateTime(user.last_login)}</td>
            <td>${formatDateTime(user.created_at)}</td>
            <td>
                <button class="btn btn-secondary btn-sm" onclick="editUser(${user.id})">编辑</button>
                <button class="btn btn-danger btn-sm" onclick="deleteUser(${user.id})">删除</button>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

// 更新分页信息
function updatePagination() {
    document.getElementById('userPageInfo').textContent = `第 ${currentPage} 页`;
    document.getElementById('userPrevPage').disabled = currentPage <= 1;
    document.getElementById('userNextPage').disabled = currentPage >= totalPages;
}

// 显示用户模态框
function showUserModal(user = null) {
    const modal = document.getElementById('userFormModal');
    const form = document.getElementById('userForm');
    const title = document.getElementById('userModalTitle');

    // 重置表单
    form.reset();

    if (user) {
        title.textContent = '编辑用户';
        document.getElementById('userId').value = user.id;
        document.getElementById('userUsername').value = user.username;
        document.getElementById('userFullName').value = user.name;
        document.getElementById('userEmail').value = user.email;
        document.getElementById('userRole').value = user.role;
        document.getElementById('userStatus').value = user.is_active ? '1' : '0';
    } else {
        title.textContent = '添加用户';
        document.getElementById('userId').value = '';
        // 设置默认值
        document.getElementById('userStatus').value = '1';
        document.getElementById('userRole').value = 'user';
    }

    modal.style.display = 'block';
}

// 保存用户
async function saveUser() {
    const userId = document.getElementById('userId').value;
    const isEdit = !!userId;

    // 获取表单数据
    const userData = {
        username: document.getElementById('userUsername').value,
        password: document.getElementById('userPassword').value,
        name: document.getElementById('userFullName').value,
        email: document.getElementById('userEmail').value,
        role: document.getElementById('userRole').value,
        is_active: document.getElementById('userStatus').value === '1'
    };

    // 表单验证
    if (!userData.username) {
        showMessage('用户名不能为空', 'error');
        return;
    }
    if (!isEdit && !userData.password) {
        showMessage('密码不能为空', 'error');
        return;
    }
    if (!userData.name) {
        showMessage('姓名不能为空', 'error');
        return;
    }
    if (!userData.email) {
        showMessage('邮箱不能为空', 'error');
        return;
    }

    // 如果是编辑模式且没有输入密码，则不发送密码字段
    if (isEdit && !userData.password) {
        delete userData.password;
    }

    try {
        const response = await fetch(isEdit ? `/api/users/${userId}` : '/api/users', {
            method: isEdit ? 'PUT' : 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(userData)
        });

        const data = await response.json();

        if (data.error) {
            showMessage(data.error, 'error');
            return;
        }

        document.getElementById('userFormModal').style.display = 'none';
        loadUsers();
        showMessage(isEdit ? '更新成功' : '创建成功', 'success');
    } catch (error) {
        showMessage(isEdit ? '更新用户失败' : '创建用户失败', 'error');
    }
}

// 编辑用户
async function editUser(id) {
    try {
        const response = await fetch(`/api/users/${id}`);
        const data = await response.json();

        if (data.error) {
            showMessage(data.error, 'error');
            return;
        }

        showUserModal(data.user);
    } catch (error) {
        showMessage('获取用户信息失败', 'error');
    }
}

// 删除用户
async function deleteUser(id) {
    if (!confirm('确定要删除这个用户吗？')) {
        return;
    }

    try {
        const response = await fetch(`/api/users/${id}`, {
            method: 'DELETE'
        });

        const data = await response.json();

        if (data.error) {
            showMessage(data.error, 'error');
            return;
        }

        loadUsers();
        showMessage('删除成功', 'success');
    } catch (error) {
        showMessage('删除用户失败', 'error');
    }
}

// 格式化日期时间
function formatDateTime(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

// 显示消息提示
function showMessage(message, type = 'info') {
    alert(message); // 临时使用 alert，后续可以改为更友好的提示
} 
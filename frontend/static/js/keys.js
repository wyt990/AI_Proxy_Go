document.addEventListener('DOMContentLoaded', function() {
    const keyTableBody = document.querySelector('#keysList');
    const addKeyBtn = document.getElementById('addKeyBtn');
    const pageSizeSelect = document.getElementById('pageSize');
    const prevPageBtn = document.getElementById('prevPage');
    const nextPageBtn = document.getElementById('nextPage');
    const currentPageSpan = document.getElementById('currentPage');
    const keyModal = document.getElementById('keyModal');
    const providerSelect = document.getElementById('providerId');

    // 从 localStorage 获取用户信息
    const user = JSON.parse(localStorage.getItem('user') || '{}');
    const isAdmin = user.role === 'admin'; // 使用 role 字段判断是否为管理员
    
    // 添加调试日志
    //console.log('User info:', user);
    //console.log('User role:', user.role);
    //console.log('Is admin:', isAdmin);

    let currentPage = 1;
    let pageSize = parseInt(pageSizeSelect.value);

    // Fetch and display keys
    function loadKeys() {
        fetch(`/api/keys?page=${currentPage}&pageSize=${pageSize}`)
            .then(response => response.json())
            .then(data => {
                keyTableBody.innerHTML = '';
                data.items.forEach(key => {
                    const row = document.createElement('tr');
                    row.innerHTML = `
                        <td>${key.ID}</td>
                        <td>${key.Name}</td>
                        <td>${key.Type === 'PUBLIC' ? '公钥' : '私钥'}</td>
                        <td>${key.UserID === 0 ? '公共' : key.UserID}</td>
                        <td>${key.ProviderName ? `${key.ProviderID} (${key.ProviderName})` : key.ProviderID}</td>
                        <td>${key.IsActive ? '启用' : '禁用'}</td>
                        <td>${key.CreatorName}</td>
                        <td>
                            <button onclick="editKey(${key.ID})" class="btn btn-edit">编辑</button>
                            <button onclick="deleteKey(${key.ID})" class="btn btn-delete">删除</button>
                        </td>
                    `;
                    keyTableBody.appendChild(row);
                });
                currentPageSpan.textContent = currentPage;
            });
    }

    // 加载服务商列表
    function loadProviders() {
        return fetch('/api/providers')
            .then(response => {
                if (!response.ok) {
                    throw new Error('获取服务商列表失败');
                }
                return response.json();
            })
            .then(data => {
                //console.log('获取到的服务商数据:', data);  // 添加日志
                
                const providerSelect = document.getElementById('providerId');
                providerSelect.innerHTML = '<option value="">请选择服务商</option>';
                
                if (data.items && Array.isArray(data.items)) {
                    data.items.forEach(provider => {
                        // 注意：使用大写字母开头的属性名，因为后端返回的是大写开头
                        const option = document.createElement('option');
                        option.value = provider.ID;  // 改为大写的 ID
                        option.textContent = `${provider.Name} (${provider.Type})`; // 改为大写的 Name
                        providerSelect.appendChild(option);
                    });
                }
                
                //console.log('服务商列表设置完成:', {
                //    providers: data.items,
                //    selectOptions: Array.from(providerSelect.options).map(opt => ({
                //        value: opt.value,
                //        text: opt.text
                //    }))
                //});
            })
            .catch(error => {
                console.error('加载服务商列表失败:', error);
                alert('加载服务商列表失败，请刷新页面重试');
            });
    }

    // 显示模态框
    function showModal(title = '添加密钥') {
        document.getElementById('modalTitle').textContent = title;
        document.getElementById('keyForm').reset();
        
        // 设置默认过期时间（3年后）
        const defaultExpireTime = new Date();
        defaultExpireTime.setFullYear(defaultExpireTime.getFullYear() + 3);
        document.getElementById('expireTime').value = defaultExpireTime.toISOString().slice(0, 16);
        
        // 根据用户角色设置类型选择框
        const typeSelect = document.getElementById('type');
        typeSelect.innerHTML = ''; // 清空现有选项
        
        // 如果是管理员角色，显示所有选项
        if (isAdmin) {
            typeSelect.innerHTML = `
                <option value="public">公钥</option>
                <option value="private">私钥</option>
            `;
        } else {
            // 非管理员只能选择私钥
            typeSelect.innerHTML = `
                <option value="private">私钥</option>
            `;
        }

        loadProviders(); // 加载服务商列表
        keyModal.style.display = 'block';
    }

    // 关闭模态框
    window.closeModal = function() {
        keyModal.style.display = 'none';
    }

    // 点击模态框外部关闭
    window.onclick = function(event) {
        if (event.target == keyModal) {
            closeModal();
        }
    }

    // Add key event
    addKeyBtn.addEventListener('click', function() {
        showModal();
    });

    // Page size change event
    pageSizeSelect.addEventListener('change', function() {
        pageSize = parseInt(this.value);
        currentPage = 1; // Reset to first page
        loadKeys();
    });

    // Pagination events
    prevPageBtn.addEventListener('click', function() {
        if (currentPage > 1) {
            currentPage--;
            loadKeys();
        }
    });

    nextPageBtn.addEventListener('click', function() {
        currentPage++;
        loadKeys();
    });

    // Save key
    window.saveKey = function() {
        const expireTimeStr = document.getElementById('expireTime').value;
        // 从localStorage获取当前用户信息
        const currentUser = JSON.parse(localStorage.getItem('user') || '{}');
        
        const keyData = {
            Name: document.getElementById('name').value.trim(),
            Type: document.getElementById('type').value.toUpperCase(),
            KeyValue: document.getElementById('keyValue').value.trim(),
            ProviderID: parseInt(document.getElementById('providerId').value),
            IsActive: document.getElementById('status').value === 'active',
            UserID: document.getElementById('type').value.toUpperCase() === 'PUBLIC' ? 0 : null,
            RateLimit: parseInt(document.getElementById('rateLimit').value) || 0,
            QuotaLimit: parseInt(document.getElementById('quotaLimit').value) || 0,
            ExpireTime: expireTimeStr ? new Date(expireTimeStr).toISOString() : null,
            CreatorID: currentUser.id || 0,
            CreatorName: currentUser.name || '未知用户'
        };

        // 验证表单
        if (!keyData.Name) {
            alert('请输入密钥名称');
            return;
        }
        if (!keyData.ProviderID) {
            alert('请选择服务商');
            return;
        }
        if (!keyData.KeyValue) {
            alert('请输入密钥值');
            return;
        }

        // 非管理员不能创建公钥
        if (!isAdmin && keyData.Type === 'PUBLIC') {
            alert('您没有权限创建公钥');
            return;
        }

        const keyId = document.getElementById('keyId').value;

        // 先检查密钥是否已存在
        fetch(`/api/keys/check?key=${encodeURIComponent(keyData.KeyValue)}`)
            .then(response => response.json())
            .then(data => {
                if (data.exists && (!keyId || data.id != keyId)) {
                    throw new Error('该密钥已存在，请勿重复添加');
                }

                // 密钥不存在，继续保存
                const method = keyId ? 'PUT' : 'POST';
                const url = keyId ? `/api/keys/${keyId}` : '/api/keys';

                return fetch(url, {
                    method: method,
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(keyData)
                });
            })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(errorData => {
                        throw new Error(errorData.error || '保存失败');
                    });
                }
                return response.json();
            })
            .then(data => {
                closeModal();
                loadKeys();
            })
            .catch(error => {
                console.error('Error:', error);
                alert(error.message || '保存失败，请重试');
            });
    };

    // Edit key
    window.editKey = function(keyId) {
        //console.log('开始编辑密钥 - ID:', keyId);
        
        // 先加载服务商列表
        loadProviders()
            .then(() => {
                // 再获取密钥详情
                return fetch(`/api/keys/${keyId}`);
            })
            .then(response => {
                //console.log('API响应状态:', response.status);
                return response.json();
            })
            .then(key => {
                //console.log('获取到的密钥数据:', key);
                
                // 记录每个字段的填充
                const fields = ['keyId', 'name', 'type', 'providerId', 'keyValue', 'rateLimit', 'quotaLimit', 'expireTime', 'status'];
                fields.forEach(field => {
                    const element = document.getElementById(field);
                    //console.log(`设置 ${field}:`, {
                    //    value: key[field],
                    //    element: element,
                    //    exists: !!element
                    //});
                });

                // 填充表单
                document.getElementById('keyId').value = key.ID;
                document.getElementById('name').value = key.Name;
                document.getElementById('type').value = key.Type;
                document.getElementById('providerId').value = key.ProviderID;
                document.getElementById('status').value = key.IsActive ? 'active' : 'inactive';
                document.getElementById('keyValue').value = key.KeyValue;
                document.getElementById('rateLimit').value = key.RateLimit || 0;
                document.getElementById('quotaLimit').value = key.QuotaLimit || 0;
                
                if (key.ExpireTime) {
                    // 将时间戳转换为本地时间字符串
                    const expireTime = new Date(key.ExpireTime).toISOString().slice(0, 16);
                    document.getElementById('expireTime').value = expireTime;
                } else {
                    document.getElementById('expireTime').value = '';
                }

                // 修改模态框标题
                document.getElementById('modalTitle').textContent = '编辑密钥';

                // 显示模态框
                const modal = document.getElementById('keyModal');
                //console.log('模态框元素:', modal);
                modal.style.display = 'block';

                // 检查服务商是否正确设置
                const providerSelect = document.getElementById('providerId');
                //console.log('服务商选择状态:', {
                //    selectedValue: providerSelect.value,
                //    expectedValue: key.ProviderID,
                //    options: Array.from(providerSelect.options).map(opt => ({
                //    value: opt.value,
                //    text: opt.text
                //    }))
                //});
            })
            .catch(error => {
                //console.error('编辑密钥失败:', error);
                alert('获取密钥详情失败');
            });
    };

    // Delete key
    window.deleteKey = function(keyId) {
        if (confirm('确定要删除这个密钥吗？')) {
            fetch(`/api/keys/${keyId}`, {
                method: 'DELETE'
            })
            .then(response => response.json())
            .then(data => {
                loadKeys();
            })
            .catch(error => {
                //console.error('Error:', error);
                alert('删除失败，请重试');
            });
        }
    };

    // Load keys on page load
    loadKeys();
}); 
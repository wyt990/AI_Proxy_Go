// 当前页码和每页显示数量
let currentPage = 1;
let pageSize = 20;

// 服务商类型预设配置
const providerPresets = {
    'OPENAI': {
        baseURL: 'https://api.openai.com/v1',
        requestFormat: JSON.stringify({
            type: 'object',
            required: ['messages', 'model'],
            properties: {
                messages: {
                    type: 'array',
                    items: {
                        type: 'object',
                        required: ['role', 'content'],
                        properties: {
                            role: { type: 'string', enum: ['system', 'user', 'assistant'] },
                            content: { type: 'string' }
                        }
                    }
                },
                model: { type: 'string' },
                temperature: { type: 'number', minimum: 0, maximum: 2 },
                max_tokens: { type: 'integer', minimum: 1 }
            }
        }, null, 2),
        responseFormat: JSON.stringify({
            type: 'object',
            required: ['choices'],
            properties: {
                id: { type: 'string' },
                object: { type: 'string' },
                created: { type: 'integer' },
                model: { type: 'string' },
                choices: {
                    type: 'array',
                    items: {
                        type: 'object',
                        properties: {
                            index: { type: 'integer' },
                            message: {
                                type: 'object',
                                properties: {
                                    role: { type: 'string' },
                                    content: { type: 'string' }
                                }
                            },
                            finish_reason: { type: 'string' }
                        }
                    }
                },
                usage: {
                    type: 'object',
                    properties: {
                        prompt_tokens: { type: 'integer' },
                        completion_tokens: { type: 'integer' },
                        total_tokens: { type: 'integer' }
                    }
                }
            }
        }, null, 2),
        authFormat: JSON.stringify({
            type: 'object',
            required: ['api_key'],
            properties: {
                api_key: { type: 'string', title: 'API Key' },
                organization: { type: 'string', title: 'Organization ID' }
            }
        }, null, 2)
    },

    'OPENAI_COMPATIBLE': {
        baseURL: 'http://localhost:8000/v1',
        requestFormat: JSON.stringify({
            type: 'object',
            required: ['messages'],
            properties: {
                messages: {
                    type: 'array',
                    items: {
                        type: 'object',
                        required: ['role', 'content'],
                        properties: {
                            role: { type: 'string', enum: ['system', 'user', 'assistant'] },
                            content: { type: 'string' }
                        }
                    }
                },
                model: { type: 'string' },
                temperature: { type: 'number', minimum: 0, maximum: 2 },
                max_tokens: { type: 'integer', minimum: 1 }
            }
        }, null, 2),
        responseFormat: JSON.stringify({
            type: 'object',
            required: ['choices'],
            properties: {
                id: { type: 'string' },
                object: { type: 'string' },
                created: { type: 'integer' },
                model: { type: 'string' },
                choices: {
                    type: 'array',
                    items: {
                        type: 'object',
                        properties: {
                            index: { type: 'integer' },
                            message: {
                                type: 'object',
                                properties: {
                                    role: { type: 'string' },
                                    content: { type: 'string' }
                                }
                            },
                            finish_reason: { type: 'string' }
                        }
                    }
                }
            }
        }, null, 2),
        authFormat: JSON.stringify({
            type: 'object',
            required: ['api_key'],
            properties: {
                api_key: { type: 'string', title: 'API Key' }
            }
        }, null, 2)
    },

    'ANTHROPIC': {
        baseURL: 'https://api.anthropic.com/v1',
        requestFormat: JSON.stringify({
            type: 'object',
            required: ['messages', 'model'],
            properties: {
                messages: {
                    type: 'array',
                    items: {
                        type: 'object',
                        required: ['role', 'content'],
                        properties: {
                            role: { type: 'string', enum: ['user', 'assistant'] },
                            content: { type: 'string' }
                        }
                    }
                },
                model: { type: 'string' },
                max_tokens: { type: 'integer' },
                temperature: { type: 'number' }
            }
        }, null, 2),
        responseFormat: JSON.stringify({
            type: 'object',
            required: ['content'],
            properties: {
                id: { type: 'string' },
                type: { type: 'string' },
                role: { type: 'string' },
                content: { type: 'string' },
                model: { type: 'string' },
                stop_reason: { type: 'string' },
                usage: {
                    type: 'object',
                    properties: {
                        input_tokens: { type: 'integer' },
                        output_tokens: { type: 'integer' }
                    }
                }
            }
        }, null, 2),
        authFormat: JSON.stringify({
            type: 'object',
            required: ['api_key'],
            properties: {
                api_key: { 
                    type: 'string',
                    title: 'API Key',
                    description: 'Anthropic API Key (以 sk-ant- 开头)'
                }
            }
        }, null, 2)
    },

    'GoogleGemini': {
        baseURL: 'https://generativelanguage.googleapis.com/v1',
        requestFormat: JSON.stringify({
            type: 'object',
            required: ['contents'],
            properties: {
                contents: {
                    type: 'array',
                    items: {
                        type: 'object',
                        required: ['role', 'parts'],
                        properties: {
                            role: { type: 'string', enum: ['user', 'model'] },
                            parts: {
                                type: 'array',
                                items: {
                                    type: 'object',
                                    required: ['text'],
                                    properties: {
                                        text: { type: 'string' }
                                    }
                                }
                            }
                        }
                    }
                },
                generationConfig: {
                    type: 'object',
                    properties: {
                        temperature: { type: 'number' },
                        maxOutputTokens: { type: 'integer' }
                    }
                }
            }
        }, null, 2),
        responseFormat: JSON.stringify({
            type: 'object',
            required: ['candidates'],
            properties: {
                candidates: {
                    type: 'array',
                    items: {
                        type: 'object',
                        properties: {
                            content: {
                                type: 'object',
                                properties: {
                                    role: { type: 'string' },
                                    parts: {
                                        type: 'array',
                                        items: {
                                            type: 'object',
                                            properties: {
                                                text: { type: 'string' }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }, null, 2),
        authFormat: JSON.stringify({
            type: 'object',
            required: ['api_key'],
            properties: {
                api_key: { type: 'string', title: 'API Key' }
            }
        }, null, 2)
    },

    'BAIDU': {
        baseURL: 'https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop',
        requestFormat: JSON.stringify({
            type: 'object',
            required: ['messages'],
            properties: {
                messages: {
                    type: 'array',
                    items: {
                        type: 'object',
                        required: ['role', 'content'],
                        properties: {
                            role: { type: 'string', enum: ['user', 'assistant'] },
                            content: { type: 'string' }
                        }
                    }
                },
                temperature: { type: 'number', minimum: 0, maximum: 1 },
                top_p: { type: 'number', minimum: 0, maximum: 1 }
            }
        }, null, 2),
        responseFormat: JSON.stringify({
            type: 'object',
            required: ['result'],
            properties: {
                id: { type: 'string' },
                result: { type: 'string' },
                usage: {
                    type: 'object',
                    properties: {
                        prompt_tokens: { type: 'integer' },
                        completion_tokens: { type: 'integer' },
                        total_tokens: { type: 'integer' }
                    }
                }
            }
        }, null, 2),
        authFormat: JSON.stringify({
            type: 'object',
            required: ['api_key', 'secret_key'],
            properties: {
                api_key: { type: 'string', title: 'API Key' },
                secret_key: { type: 'string', title: 'Secret Key' }
            }
        }, null, 2)
    }
};

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    // 加载服务商列表
    loadProviders();

    // 绑定事件处理器
    document.getElementById('pageSize').addEventListener('change', function() {
        pageSize = parseInt(this.value);
        currentPage = 1;
        loadProviders();
    });

    document.getElementById('addProviderBtn').addEventListener('click', showAddModal);
    document.getElementById('saveProviderBtn').addEventListener('click', saveProvider);

    // 关闭模态框
    document.querySelectorAll('.close, .close-btn').forEach(el => {
        el.addEventListener('click', closeModal);
    });

    // 当选择服务商类型时自动填充预设配置
    document.getElementById('type').addEventListener('change', function() {
        const preset = providerPresets[this.value];
        if (preset) {
            document.getElementById('baseURL').value = preset.baseURL;
            document.getElementById('requestFormat').value = preset.requestFormat;
            document.getElementById('responseFormat').value = preset.responseFormat;
            document.getElementById('authFormat').value = preset.authFormat;
        } else {
            // 如果选择"自定义"，则清空所有字段
            document.getElementById('baseURL').value = '';
            document.getElementById('requestFormat').value = '';
            document.getElementById('responseFormat').value = '';
            document.getElementById('authFormat').value = '';
        }
    });
});

// 加载服务商列表
async function loadProviders() {
    try {
        const response = await fetch(`/api/providers?page=${currentPage}&pageSize=${pageSize}`);
        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || '加载服务商列表失败');
        }

        renderProvidersList(data.items);
        renderPagination(data.total);
    } catch (error) {
        showMessage(error.message, 'error');
    }
}

// 渲染服务商列表
function renderProvidersList(providers) {
    const tbody = document.getElementById('providersList');
    tbody.innerHTML = providers.map(provider => `
        <tr>
            <td>${provider.ID || '-'}</td>
            <td>${provider.Name || '-'}</td>
            <td>${provider.Type || '-'}</td>
            <td>${provider.BaseURL || '-'}</td>
            <td>${provider.AuthType || '-'}</td>
            <td>${provider.RateLimit ? provider.RateLimit + '/分钟' : '-'}</td>
            <td>
                ${provider.RetryCount || '0'}次/${provider.RetryInterval || '0'}秒
            </td>
            <td>
                ${provider.LastCheckTime ? formatTimeAgo(provider.LastCheckTime) : '-'}
            </td>
            <td>
                <span class="status-badge ${getStatusClass(provider.Status)}">
                    ${getStatusText(provider.Status)}
                </span>
            </td>
            <td>${provider.LastError || '-'}</td>
            <td>
                <button class="btn btn-sm btn-check" onclick="checkProvider(${provider.ID})" title="健康检查">
                    <i class="fas fa-heartbeat"></i>
                </button>
                <button class="btn btn-sm btn-edit" onclick="editProvider(${provider.ID})">
                    编辑
                </button>
                <button class="btn btn-sm btn-delete" onclick="deleteProvider(${provider.ID})">
                    删除
                </button>
            </td>
        </tr>
    `).join('');
}

// 显示添加模态框
function showAddModal() {
    document.getElementById('modalTitle').textContent = '添加服务商';
    document.getElementById('providerForm').reset();
    document.getElementById('providerId').value = '';
    
    // 触发类型选择的change事件，自动填充默认值
    const typeSelect = document.getElementById('type');
    if (typeSelect.value) {
        const event = new Event('change');
        typeSelect.dispatchEvent(event);
    }

    document.getElementById('providerModal').style.display = 'block';
}

// 关闭模态框
function closeModal() {
    document.getElementById('providerModal').style.display = 'none';
}

// 保存服务商
async function saveProvider() {
    try {
        const provider = {
            // 如果是编辑模式，将ID转换为数字类型
            ID: document.getElementById('providerId').value ? 
                parseInt(document.getElementById('providerId').value) : undefined,
            Name: document.getElementById('name').value,
            Type: document.getElementById('type').value,
            BaseURL: document.getElementById('baseURL').value,
            RequestFormat: document.getElementById('requestFormat').value,
            ResponseFormat: document.getElementById('responseFormat').value,
            AuthFormat: document.getElementById('authFormat').value,
            Headers: document.getElementById('headers').value,
            AuthType: document.getElementById('authType').value,
            ProxyMode: document.getElementById('proxyMode').value,
            RateLimit: parseInt(document.getElementById('rateLimit').value),
            Timeout: parseInt(document.getElementById('timeout').value),
            RetryCount: parseInt(document.getElementById('retryCount').value),
            RetryInterval: parseInt(document.getElementById('retryInterval').value),
        };

        // 如果是新建，删除 ID 字段
        if (!provider.ID) {
            delete provider.ID;
        }

        const url = provider.ID ? `/api/providers/${provider.ID}` : '/api/providers';
        const method = provider.ID ? 'PUT' : 'POST';

        //console.log('Saving provider:', provider);

        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(provider)
        });

        const data = await response.json();
        //console.log('Server response:', data);

        if (!response.ok) {
            throw new Error(data.error || data.message || '保存失败');
        }

        showMessage('保存成功', 'success');
        closeModal();
        loadProviders();
    } catch (error) {
        console.error('Save error:', error);
        showMessage(error.message, 'error');
    }
}

// 编辑服务商
async function editProvider(id) {
    try {
        const response = await fetch(`/api/providers/${id}`);
        const provider = await response.json();

        if (!response.ok) {
            throw new Error(provider.error || '加载服务商信息失败');
        }

        document.getElementById('modalTitle').textContent = '编辑服务商';
        document.getElementById('providerId').value = provider.ID;
        document.getElementById('name').value = provider.Name;
        document.getElementById('type').value = provider.Type;
        document.getElementById('baseURL').value = provider.BaseURL;
        document.getElementById('requestFormat').value = provider.RequestFormat;
        document.getElementById('responseFormat').value = provider.ResponseFormat;
        document.getElementById('authFormat').value = provider.AuthFormat;
        document.getElementById('authType').value = provider.AuthType;
        document.getElementById('proxyMode').value = provider.ProxyMode;
        document.getElementById('rateLimit').value = provider.RateLimit;
        document.getElementById('timeout').value = provider.Timeout;
        document.getElementById('retryCount').value = provider.RetryCount;
        document.getElementById('headers').value = provider.Headers;
        document.getElementById('retryInterval').value = provider.RetryInterval;

        document.getElementById('providerModal').style.display = 'block';
    } catch (error) {
        showMessage(error.message, 'error');
    }
}

// 删除服务商
async function deleteProvider(id) {
    if (!confirm('确定要删除这个服务商吗？')) {
        return;
    }

    try {
        const response = await fetch(`/api/providers/${id}`, {
            method: 'DELETE'
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '删除失败');
        }

        showMessage('删除成功', 'success');
        loadProviders();
    } catch (error) {
        showMessage(error.message, 'error');
    }
}

// 执行健康检查
async function checkProvider(id) {
    try {
        const response = await fetch(`/api/providers/${id}/check`, {
            method: 'POST'
        });

        const data = await response.json();
        if (!response.ok) {
            throw new Error(data.error || '健康检查失败');
        }

        showMessage('健康检查完成', 'success');
        loadProviders();
    } catch (error) {
        showMessage(error.message, 'error');
    }
}

// 工具函数
function getStatusClass(status) {
    const statusMap = {
        'NORMAL': 'status-normal',
        'ERROR': 'status-error',
        'RATE_LIMITED': 'status-limited'
    };
    return statusMap[status] || '';
}

function getStatusText(status) {
    const statusMap = {
        'NORMAL': '正常',
        'ERROR': '错误',
        'RATE_LIMITED': '限流'
    };
    return statusMap[status] || status;
}

function showMessage(message, type = 'info') {
    // 实现消息提示
}

// 格式化时间差
function formatTimeAgo(timestamp) {
    const now = new Date();
    const time = new Date(timestamp);
    const diff = Math.floor((now - time) / 1000);

    if (diff < 60) return '刚刚';
    if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`;
    return `${Math.floor(diff / 86400)}天前`;
}

// 截断文本
function truncateText(text, length) {
    return text.length > length ? text.substring(0, length) + '...' : text;
} 
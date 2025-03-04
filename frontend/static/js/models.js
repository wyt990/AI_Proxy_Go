// 当前页码和每页显示数量
let currentPage = 1;
let pageSize = 20;

// 服务商模型参数模板
const modelParametersTemplates = {
    'OPENAI': {
        temperature: 0.7,
        top_p: 1,
        frequency_penalty: 0,
        presence_penalty: 0,
        stop: null
    },
    'OPENAI_COMPATIBLE': {
        temperature: 0.7,
        top_p: 1,
        frequency_penalty: 0,
        presence_penalty: 0,
        stop: null
    },
    'ANTHROPIC': {
        temperature: 0.7,
        top_k: 40,
        top_p: 1,
        max_tokens_to_sample: 4096
    },
    'GoogleGemini': {
        temperature: 0.7,
        top_p: 0.8,
        candidate_count: 1,
        stop_sequences: [],
        max_output_tokens: 2048
    },
    'BAIDU': {
        temperature: 0.8,
        top_p: 0.8,
        penalty_score: 1.0,
        system: ""  // 系统提示词
    }
};

// 添加模型类型映射
const MODEL_TYPES = {
    'chat': '对话',
    'text2image': '文生图',
    'image2image': '图生图',
    'audio': '语音',
    'video': '视频',
    'embedding': '嵌入',
    'tool': '工具'
};

// 页面加载完成后执行
document.addEventListener('DOMContentLoaded', function() {
    //console.log('页面加载完成，开始初始化');
    // 加载模型列表
    loadModels();
    
    // 绑定添加按钮事件
    document.getElementById('addModelBtn').addEventListener('click', showAddModal);
    
    // 加载服务商列表
    loadProviders();
    
    // 绑定服务商选择事件
    //console.log('开始绑定服务商选择事件');
    const providerSelect = document.getElementById('providerId');
    if (providerSelect) {
        providerSelect.addEventListener('change', handleProviderChange);
        //console.log('服务商选择事件绑定成功');
    } else {
        console.error('未找到服务商选择下拉框元素');
    }
});

// 加载服务商列表
async function loadProviders() {
    try {
        //console.log('开始加载服务商列表');
        const response = await fetch('/api/providers');
        const data = await response.json();
        //console.log('获取到的服务商列表:', data);
        
        if (!response.ok) {
            throw new Error(data.error || '加载服务商列表失败');
        }
        
        const select = document.getElementById('providerId');
        select.innerHTML = data.items.map(provider => `
            <option value="${provider.ID}">${provider.Name}</option>
        `).join('');
        //console.log('服务商下拉列表已更新');

        // 如果只有一个服务商，主动触发change事件
        if (data.items && data.items.length === 1) {
            //console.log('检测到只有一个服务商，主动触发change事件');
            select.value = data.items[0].ID;
            handleProviderChange.call(select);
        }

    } catch (error) {
        console.error('加载服务商列表失败:', error);
        showMessage(error.message, 'error');
    }
}

// 加载模型列表
async function loadModels() {
    try {
        const response = await fetch(`/api/models?page=${currentPage}&pageSize=${pageSize}`);
        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || '加载模型列表失败');
        }
        
        renderModelsList(data.items);
    } catch (error) {
        showMessage(error.message, 'error');
    }
}

// 修改渲染模型列表的函数
function renderModelsList(models) {
    const tbody = document.getElementById('modelsList');
    tbody.innerHTML = models.map(model => `
        <tr>
            <td>${model.ID}</td>
            <td>${model.Name}</td>
            <td>${model.Provider?.Name || '-'}</td>
            <td>${model.ModelID}</td>
            <td>${MODEL_TYPES[model.Type] || model.Type}</td>  <!-- 新增模型类型列 -->
            <td>
                <span class="status-badge ${getStatusClass(model.Status)}">
                    ${getStatusText(model.Status)}
                </span>
            </td>
            <td>${model.MaxTokens || '不限制'}</td>
            <td>${formatPrice(model.InputPrice)}/${formatPrice(model.OutputPrice)}</td>
            <td>${model.Sort}</td>
            <td>
                <button class="btn btn-sm btn-edit" onclick="editModel(${model.ID})">
                    编辑
                </button>
                <button class="btn btn-sm btn-delete" onclick="deleteModel(${model.ID})">
                    删除
                </button>
            </td>
        </tr>
    `).join('');
}

// 格式化价格显示
function formatPrice(price) {
    return Number(price).toFixed(6);
}

// 显示添加模态框
function showAddModal() {
    //console.log('显示添加模态框');
    document.getElementById('modalTitle').textContent = '添加模型';
    document.getElementById('modelForm').reset();
    document.getElementById('modelId').value = '';
    // 设置默认值
    document.getElementById('status').value = 'NORMAL';
    document.getElementById('maxTokens').value = '0';
    document.getElementById('inputPrice').value = '0.000000';
    document.getElementById('outputPrice').value = '0.000000';
    document.getElementById('sort').value = '0';
    document.getElementById('parameters').value = '{}';
    document.getElementById('modelModal').style.display = 'block';
    
    // 获取服务商选择框并触发change事件
    const select = document.getElementById('providerId');
    if (select && select.value) {
        //console.log('触发服务商选择事件');
        handleProviderChange.call(select);
    }
}

// 关闭模态框
function closeModal() {
    document.getElementById('modelModal').style.display = 'none';
}

// 处理服务商选择变化
async function handleProviderChange() {
    //console.log('服务商选择变更事件触发');
    try {
        const providerId = this.value;
        //console.log('选择的服务商ID:', providerId);
        if (!providerId) return;

        // 获取服务商信息
        //console.log('开始获取服务商信息...');
        const response = await fetch(`/api/providers/${providerId}`);
        const provider = await response.json();
        //console.log('获取到的服务商信息:', provider);

        if (!response.ok) {
            throw new Error(provider.error || '获取服务商信息失败');
        }

        // 根据服务商类型设置参数模板
        //console.log('服务商类型:', provider.Type);
        const template = modelParametersTemplates[provider.Type] || {};
        //console.log('对应的参数模板:', template);
        const parametersField = document.getElementById('parameters');
        //console.log('当前参数字段值:', parametersField.value);
        // 总是设置参数模板，除非是在编辑模式下
        const modelId = document.getElementById('modelId').value;
        if (!modelId) {  // 只在添加新模型时设置模板
            //console.log('设置参数模板');
            parametersField.value = JSON.stringify(template, null, 2);
        }

        // 添加参数说明
        const description = getParametersDescription(provider.Type);
        //console.log('设置参数说明:', description);
        parametersField.setAttribute('placeholder', description);
    } catch (error) {
        console.error('处理服务商选择变更时出错:', error);
        showMessage(error.message, 'error');
    }
}

// 获取参数说明
function getParametersDescription(providerType) {
    const descriptions = {
        'OPENAI': `参数说明：
- temperature: 采样温度，范围0-2，越高越随机
- top_p: 核采样，范围0-1，控制输出随机性
- frequency_penalty: 频率惩罚，范围-2.0-2.0
- presence_penalty: 存在惩罚，范围-2.0-2.0
- stop: 停止生成的标记序列，可以是字符串或字符串数组`,

        'OPENAI_COMPATIBLE': `参数说明：
- temperature: 采样温度，范围0-2，越高越随机
- top_p: 核采样，范围0-1，控制输出随机性
- frequency_penalty: 频率惩罚，范围-2.0-2.0
- presence_penalty: 存在惩罚，范围-2.0-2.0
- stop: 停止生成的标记序列，可以是字符串或字符串数组`,

        'ANTHROPIC': `参数说明：
- temperature: 采样温度，范围0-1
- top_k: 保留概率最高的k个token
- top_p: 核采样，范围0-1
- max_tokens_to_sample: 最大生成token数`,

        'GoogleGemini': `参数说明：
- temperature: 采样温度，范围0-1
- top_p: 核采样，范围0-1
- candidate_count: 生成候选数量
- stop_sequences: 停止序列
- max_output_tokens: 最大输出token数`,

        'BAIDU': `参数说明：
- temperature: 采样温度，范围0-1
- top_p: 核采样，范围0-1
- penalty_score: 重复惩罚，范围1.0-2.0
- system: 系统提示词，用于设置角色和行为`
    };

    return descriptions[providerType] || '请根据服务商API文档填写相应参数';
}

// 验证JSON格式
function validateParameters() {
    const parametersField = document.getElementById('parameters');
    try {
        if (parametersField.value) {
            // 验证并格式化 JSON
            const parsed = JSON.parse(parametersField.value);
            parametersField.value = JSON.stringify(parsed, null, 2);
        }
        return true;
    } catch (error) {
        console.error('JSON 格式验证失败:', error);
        showMessage('模型参数JSON格式无效', 'error');
        return false;
    }
}

// 修改保存函数，添加参数验证
async function saveModel() {
    if (!validateParameters()) {
        return;
    }

    try {
        // 先解析参数，确保是有效的 JSON 对象
        const parameters = JSON.parse(document.getElementById('parameters').value.trim() || '{}');
        
        const model = {
            ID: parseInt(document.getElementById('modelId').value) || undefined,
            Name: document.getElementById('name').value.trim(),
            ProviderID: parseInt(document.getElementById('providerId').value),
            ModelID: document.getElementById('ModelID').value.trim(),
            Status: document.getElementById('status').value,
            MaxTokens: parseInt(document.getElementById('maxTokens').value),
            InputPrice: formatPrice(document.getElementById('inputPrice').value),
            OutputPrice: formatPrice(document.getElementById('outputPrice').value),
            Description: document.getElementById('description').value.trim(),
            Parameters: parameters,
            Sort: parseInt(document.getElementById('sort').value),
        };

        //console.log('保存模型数据:', model);

        // 如果是新建，删除ID字段
        if (!model.ID) {
            delete model.ID;
        }

        const url = model.ID ? `/api/models/${model.ID}` : '/api/models';
        const method = model.ID ? 'PUT' : 'POST';

        // 确保 Parameters 是字符串格式
        const requestData = {
            ...model,
            Parameters: JSON.stringify(model.Parameters)
        };

        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(requestData)
        });

        console.log('发送到服务器的数据:', JSON.stringify(requestData, null, 2));

        if (!response.ok) {
            const errorData = await response.json();
            console.error('保存失败:', errorData);
            // 尝试获取更详细的错误信息
            const errorMessage = errorData.error || 
                               errorData.message || 
                               errorData.details || 
                               '保存失败';
            throw new Error(errorMessage);
        }

        const data = await response.json();

        showMessage('保存成功', 'success');
        closeModal();
        loadModels();
    } catch (error) {
        console.error('保存模型出错:', error);
        showMessage(error.message, 'error');
    }
}

// 编辑模型
async function editModel(id) {
    try {
        const response = await fetch(`/api/models/${id}`);
        const model = await response.json();

        if (!response.ok) {
            throw new Error(model.error || '加载模型信息失败');
        }

        document.getElementById('modalTitle').textContent = '编辑模型';
        document.getElementById('modelId').value = model.ID;
        document.getElementById('name').value = model.Name;
        document.getElementById('providerId').value = model.ProviderID;
        document.getElementById('ModelID').value = model.ModelID;
        document.getElementById('status').value = model.Status;
        document.getElementById('maxTokens').value = model.MaxTokens;
        document.getElementById('inputPrice').value = Number(model.InputPrice).toFixed(6);
        document.getElementById('outputPrice').value = Number(model.OutputPrice).toFixed(6);
        document.getElementById('description').value = model.Description;
        document.getElementById('parameters').value = model.Parameters;
        document.getElementById('sort').value = model.Sort;

        // 触发服务商变更事件以加载参数模板
        handleProviderChange.call(document.getElementById('providerId'));

        document.getElementById('modelModal').style.display = 'block';
    } catch (error) {
        showMessage(error.message, 'error');
    }
}

// 删除模型
async function deleteModel(id) {
    if (!confirm('确定要删除这个模型吗？')) {
        return;
    }

    try {
        const response = await fetch(`/api/models/${id}`, {
            method: 'DELETE'
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '删除失败');
        }

        showMessage('删除成功', 'success');
        loadModels();
    } catch (error) {
        showMessage(error.message, 'error');
    }
}

// 工具函数
function getStatusClass(status) {
    return status === 'NORMAL' ? 'status-normal' : 'status-disabled';
}

function getStatusText(status) {
    return status === 'NORMAL' ? '正常' : '禁用';
}

function showMessage(message, type = 'info') {
    // 实现消息提示
} 
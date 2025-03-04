// 在文件开头添加消息模板
const MESSAGE_TEMPLATE = {
    assistant: (content) => {
        // 检查是否包含思维过程和最终回答
        if (content.includes('思维过程：') && content.includes('最终回答：')) {
            // 分离思维过程和最终回答
            const thinkingMatch = content.match(/思维过程：([\s\S]*?)最终回答：([\s\S]*)/);
            if (thinkingMatch) {
                const [_, thinking, answer] = thinkingMatch;
                return `
                    <div class="message-wrapper">
                        <div class="message-content markdown-body">
                            <div class="thinking-process">
                                <strong>思维过程：</strong>
                                ${thinking.trim()}
                            </div>
                            <div class="final-answer-container">
                                <div class="final-answer">
                                    <strong>最终回答：</strong>
                                    ${answer.trim()}
                                </div>
                                <div class="message-actions">
                                    <button class="action-btn copy-btn" title="复制回答">
                                        <i class="fas fa-copy"></i>
                                    </button>
                                    <button class="action-btn delete-btn" title="删除消息">
                                        <i class="fas fa-trash"></i>
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                `;
            }
        }
        
        // 如果没有特殊格式，使用默认模板
        return `
            <div class="message-wrapper">
                <div class="message-content markdown-body">
                    ${content}
                    <div class="message-actions">
                        <button class="action-btn copy-btn" title="复制消息">
                            <i class="fas fa-copy"></i>
                        </button>
                        <button class="action-btn delete-btn" title="删除消息">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </div>
            </div>
        `;
    }
};

document.addEventListener('DOMContentLoaded', function() {
    const providerSelect = document.getElementById('providerId');
    const modelSelect = document.getElementById('modelId');
    const keySelect = document.getElementById('keyId');
    const messageInput = document.getElementById('messageInput');
    const sendButton = document.getElementById('sendButton');
    const chatMessages = document.getElementById('chatMessages');
    const newSessionBtn = document.getElementById('newSessionBtn');

    // 从localStorage获取用户信息
    const currentUser = JSON.parse(localStorage.getItem('user') || '{}');
    //console.log('User info:', currentUser);
    console.log('User id:', currentUser.id);


    // 添加新的变量
    let currentSessionId = null;

    // 添加全局变量
    let internetEnabled = false;
    let deepThoughtEnabled = false;
    const internetToggle = document.getElementById('internetToggle');
    const deepThoughtToggle = document.getElementById('deepThoughtToggle');
    const featureToggles = document.querySelector('.feature-toggles');

    // 加载服务商列表
    function loadProviders() {
        console.log('开始加载服务商列表');
        fetch('/api/chat/providers', {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        })
            .then(response => response.json())
            .then(data => {
                providerSelect.innerHTML = '<option value="">请选择服务商</option>';
                data.items.forEach(provider => {
                    providerSelect.innerHTML += `
                        <option value="${provider.id}">${provider.name}</option>
                    `;
                });
            })
            
            .catch(error => {
                //console.error('加载服务商列表失败:', error);
                providerSelect.innerHTML = '<option value="">加载失败，请刷新重试</option>';
            });
    }

    // 加载模型列表
    function loadModels(providerId) {
        //console.log('开始加载模型列表，服务商ID:', providerId);
        modelSelect.disabled = !providerId;
        if (!providerId) {
            modelSelect.innerHTML = '<option value="">请先选择服务商</option>';
            return Promise.resolve();
        }

        return fetch(`/api/chat/provider/${providerId}/models`,{
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        })
            .then(response => response.json())
            .then(data => {
                //console.log('收到模型数据:', data);
                modelSelect.innerHTML = '<option value="">请选择模型</option>';
                if (data.items && Array.isArray(data.items)) {
                    data.items.forEach(model => {
                        //console.log('处理模型:', model);
                        const option = document.createElement('option');
                        option.value = model.id;
                        option.textContent = model.name;
                        if (model.parameters) {
                            //console.log('模型参数(原始):', model.parameters);
                            option.dataset.parameters = typeof model.parameters === 'string' 
                                ? model.parameters 
                                : JSON.stringify(model.parameters);
                            //console.log('模型参数(处理后):', option.dataset.parameters);
                        } else {
                            //console.log('模型没有参数');
                        }
                        modelSelect.appendChild(option);
                    });
                }
                modelSelect.disabled = false;
            });
    }

    // 添加参数检查函数
    function checkModelParameters() {
        const selectedOption = modelSelect.options[modelSelect.selectedIndex];
        //console.log('检查选中的模型:', selectedOption);
        
        if (!selectedOption) {
            //console.log('没有选中的模型，隐藏按钮');
            internetToggle.style.display = 'none';
            deepThoughtToggle.style.display = 'none';
            return;
        }

        //console.log('模型参数数据:', selectedOption.dataset.parameters);

        let params = {};
        if (selectedOption.dataset.parameters) {
            try {
                params = JSON.parse(selectedOption.dataset.parameters);
                //console.log('解析后的参数:', params);
            } catch (e) {
                //console.error('解析模型参数失败:', e);
                params = {};
            }
        }

        // 分别检查每个功能参数
        //console.log('检查功能参数:', {
        //    enable_internet: params.enable_internet,
        //    enable_deep_thought: params.enable_deep_thought
        //});

        if (params.enable_internet !== undefined) {
            //console.log('显示联网按钮，状态:', params.enable_internet);
            internetToggle.style.display = 'block';
            internetEnabled = !!params.enable_internet;
            internetToggle.classList.toggle('active', internetEnabled);
        } else {
            //console.log('隐藏联网按钮');
            internetToggle.style.display = 'none';
        }

        if (params.enable_deep_thought !== undefined) {
            //console.log('显示深度思考按钮，状态:', params.enable_deep_thought);
            deepThoughtToggle.style.display = 'block';
            deepThoughtEnabled = !!params.enable_deep_thought;
            deepThoughtToggle.classList.toggle('active', deepThoughtEnabled);
        } else {
            //console.log('隐藏深度思考按钮');
            deepThoughtToggle.style.display = 'none';
        }
    }

    // 更新按钮状态（只更新显示的按钮）
    function updateToggleButtons() {
        if (internetToggle.style.display !== 'none') {
            internetToggle.classList.toggle('active', internetEnabled);
        }
        if (deepThoughtToggle.style.display !== 'none') {
            deepThoughtToggle.classList.toggle('active', deepThoughtEnabled);
        }
    }

    // 添加按钮点击事件
    internetToggle.addEventListener('click', () => {
        internetEnabled = !internetEnabled;
        updateToggleButtons();
    });

    deepThoughtToggle.addEventListener('click', () => {
        deepThoughtEnabled = !deepThoughtEnabled;
        updateToggleButtons();
    });

    // 加载密钥列表
    function loadKeys(providerId) {
        keySelect.disabled = !providerId;
        if (!providerId) {
            keySelect.innerHTML = '<option value="">请先选择服务商</option>';
            return Promise.resolve();
        }
        console.log('开始加载密钥列表，服务商ID:', providerId);
        return fetch(`/api/chat/provider/${providerId}/keys`)
            .then(response => response.json())
            .then(data => {
                keySelect.innerHTML = '<option value="">请选择密钥</option>';
                
                // 处理公钥
                if (data.publicKeys && data.publicKeys.length > 0) {
                    const publicKeysGroup = document.createElement('optgroup');
                    publicKeysGroup.label = '公钥';
                    data.publicKeys.forEach(key => {
                        const option = new Option(key.Name, key.ID);
                        publicKeysGroup.appendChild(option);
                    });
                    keySelect.appendChild(publicKeysGroup);
                }

                // 处理私钥
                if (data.privateKeys && data.privateKeys.length > 0) {
                    const privateKeysGroup = document.createElement('optgroup');
                    privateKeysGroup.label = '私钥';
                    data.privateKeys.forEach(key => {
                        const option = new Option(key.Name, key.ID);
                        privateKeysGroup.appendChild(option);
                    });
                    keySelect.appendChild(privateKeysGroup);
                }

                keySelect.disabled = false;
            });
    }

    // 修改添加消息函数
    function addMessage(content, role, failed = false) {
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${role} ${failed ? 'failed' : ''}`;
        
        //console.log('添加消息 - 原始数据:', {
        //    content,
        //    role,
        //    failed,
        //    contentType: typeof content,
        //    isObject: typeof content === 'object',
        //    hasId: content?.id
        //});
        
        // 设置消息ID和内容
        let messageContent;
        if (typeof content === 'object') {
            messageDiv.dataset.messageId = content.id;
            messageContent = content.content;
            //console.log('从对象提取数据:', {
            //    id: content.id,
            //    content: messageContent,
            //    setId: messageDiv.dataset.messageId
            //});
        } else {
            messageContent = content;
            //console.log('使用字符串内容:', {
            //    content: messageContent,
            //    noId: true
            //});
        }

        // 处理消息内容，保持换行和格式
        const formattedContent = formatMessageContent(messageContent);

        if (role === 'user') {
            // 保存原始内容，用于重发
            messageDiv.dataset.content = messageContent;
            messageDiv.innerHTML = `
                <div class="message-wrapper">
                    <div class="message-content markdown-body">${formattedContent}</div>
                    <div class="message-actions">
                        <button class="action-btn copy-btn" title="复制消息">
                            <i class="fas fa-copy"></i>
                        </button>
                        <button class="action-btn resend-btn" title="重新发送">
                            <i class="fas fa-redo-alt"></i>
                        </button>
                        <button class="action-btn delete-btn" title="删除消息">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </div>
            `;

            // 添加鼠标事件监听器
            messageDiv.addEventListener('mouseenter', () => {
                //console.log('鼠标进入消息区域');
                const actionsDiv = messageDiv.querySelector('.message-actions');
                //console.log('重发按钮容器:', actionsDiv);
                //console.log('重发按钮容器可见性:', getComputedStyle(actionsDiv).visibility);
                //console.log('重发按钮容器透明度:', getComputedStyle(actionsDiv).opacity);
            });

            messageDiv.addEventListener('mouseleave', () => {
                //console.log('鼠标离开消息区域');
                const actionsDiv = messageDiv.querySelector('.message-actions');
                //console.log('重发按钮容器可见性:', getComputedStyle(actionsDiv).visibility);
                //console.log('重发按钮容器透明度:', getComputedStyle(actionsDiv).opacity);
            });

            // 添加按钮事件
            const copyBtn = messageDiv.querySelector('.copy-btn');
            const resendBtn = messageDiv.querySelector('.resend-btn');
            const deleteBtn = messageDiv.querySelector('.delete-btn');

            // 复制按钮事件
            copyBtn.addEventListener('click', () => {
                const finalAnswer = messageDiv.querySelector('.final-answer');
                const content = finalAnswer ? finalAnswer.textContent.replace('最终回答：', '').trim() : messageContent;
                
                const textarea = document.createElement('textarea');
                textarea.value = content;
                textarea.style.position = 'fixed';
                textarea.style.opacity = '0';
                document.body.appendChild(textarea);
                
                try {
                    textarea.select();
                    document.execCommand('copy');
                    copyBtn.setAttribute('data-original-title', copyBtn.getAttribute('title'));
                    copyBtn.setAttribute('title', '已复制!');
                    setTimeout(() => {
                        copyBtn.setAttribute('title', copyBtn.getAttribute('data-original-title'));
                    }, 1000);
                } catch (err) {
                    console.error('复制失败:', err);
                } finally {
                    document.body.removeChild(textarea);
                }
            });

            // 删除按钮事件
            deleteBtn.addEventListener('click', async () => {
                if (confirm('确定要删除这条消息吗？')) {
                    const messageElement = messageDiv.closest('.message');
                    const messageId = messageElement.dataset.messageId;
                    
                    // 添加调试日志
                    //console.log('准备删除消息:', {
                    //    messageId: messageId,
                    //    userId: currentUser.id,
                    //    element: messageElement
                    //});
                    
                    if (!messageId) {
                        //console.error('消息ID未找到');
                        alert('无法删除消息：消息ID未找到');
                        return;
                    }

                    try {
                        // 调用后端API删除消息
                        const response = await fetch(`/api/chat/messages/${messageId}?userId=${currentUser.id}`, {
                            method: 'DELETE',
                            headers: {
                                'Authorization': `Bearer ${localStorage.getItem('token')}`
                            }
                        });

                        //console.log('删除消息响应:', {
                        //    status: response.status,
                        //    ok: response.ok
                        //});

                        if (!response.ok) {
                            const error = await response.json();
                            throw new Error(error.error || '删除失败');
                        }

                        // 如果有关联的回复，也一起删除
                        const nextMessage = messageDiv.nextElementSibling;
                        if (nextMessage && nextMessage.classList.contains('assistant')) {
                            const nextMessageId = nextMessage.dataset.messageId;
                            // 删除关联的回复消息
                            await fetch(`/api/chat/messages/${nextMessageId}?userId=${currentUser.id}`, {
                                method: 'DELETE',
                                headers: {
                                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                                }
                            });
                            nextMessage.remove();
                        }
                        
                        // 从页面中移除消息元素
                        messageElement.remove();
                    } catch (error) {
                        //console.error('删除消息失败:', error);
                        alert('删除消息失败: ' + error.message);
                    }
                }
            });

            // 添加重发按钮事件
            resendBtn.addEventListener('click', () => {
                // 移除之前的错误响应（如果有）
                const nextMessage = messageDiv.nextElementSibling;
                if (nextMessage && nextMessage.classList.contains('assistant')) {
                    nextMessage.remove();
                }
                
                // 重新发送消息
                sendMessage(messageDiv.dataset.content);
            });
        } else {
            messageDiv.innerHTML = `
                <div class="message-wrapper">
                    <div class="message-content markdown-body">${formattedContent}</div>
                    <div class="message-actions">
                        <button class="action-btn copy-btn" title="复制消息">
                            <i class="fas fa-copy"></i>
                        </button>
                        <button class="action-btn delete-btn" title="删除消息">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </div>
            `;

            // 添加按钮事件
            const copyBtn = messageDiv.querySelector('.copy-btn');
            const deleteBtn = messageDiv.querySelector('.delete-btn');

            // 复制按钮事件
            copyBtn.addEventListener('click', () => {
                const finalAnswer = messageDiv.querySelector('.final-answer');
                const content = finalAnswer ? finalAnswer.textContent.replace('最终回答：', '').trim() : messageContent;
                
                const textarea = document.createElement('textarea');
                textarea.value = content;
                textarea.style.position = 'fixed';
                textarea.style.opacity = '0';
                document.body.appendChild(textarea);
                
                try {
                    textarea.select();
                    document.execCommand('copy');
                    copyBtn.setAttribute('data-original-title', copyBtn.getAttribute('title'));
                    copyBtn.setAttribute('title', '已复制!');
                    setTimeout(() => {
                        copyBtn.setAttribute('title', copyBtn.getAttribute('data-original-title'));
                    }, 1000);
                } catch (err) {
                    console.error('复制失败:', err);
                } finally {
                    document.body.removeChild(textarea);
                }
            });

            // 删除按钮事件
            deleteBtn.addEventListener('click', async () => {
                if (confirm('确定要删除这条消息吗？')) {
                    const messageElement = messageDiv.closest('.message');
                    const messageId = messageElement.dataset.messageId;
                    
                    // 添加调试日志
                    //console.log('准备删除消息:', {
                    //    messageId: messageId,
                    //    userId: currentUser.id,
                    //    element: messageElement
                    //});
                    
                    if (!messageId) {
                        //console.error('消息ID未找到');
                        alert('无法删除消息：消息ID未找到');
                        return;
                    }

                    try {
                        // 调用后端API删除消息
                        const response = await fetch(`/api/chat/messages/${messageId}?userId=${currentUser.id}`, {
                            method: 'DELETE',
                            headers: {
                                'Authorization': `Bearer ${localStorage.getItem('token')}`
                            }
                        });

                        //console.log('删除消息响应:', {
                        //    status: response.status,
                        //    ok: response.ok
                        //});

                        if (!response.ok) {
                            const error = await response.json();
                            throw new Error(error.error || '删除失败');
                        }

                        // 如果是AI回复，同时检查是否需要删除对应的用户消息
                        const prevMessage = messageDiv.previousElementSibling;
                        if (prevMessage && prevMessage.classList.contains('user')) {
                            if (confirm('是否同时删除相关的提问消息？')) {
                                const prevMessageId = prevMessage.dataset.messageId;
                                // 删除关联的用户消息
                                await fetch(`/api/chat/messages/${prevMessageId}?userId=${currentUser.id}`, {
                                    method: 'DELETE',
                                    headers: {
                                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                                    }
                                });
                                prevMessage.remove();
                            }
                        }
                        
                        // 从页面中移除消息元素
                        messageElement.remove();
                    } catch (error) {
                        //console.error('删除消息失败:', error);
                        alert('删除消息失败: ' + error.message);
                    }
                }
            });
        }

        chatMessages.appendChild(messageDiv);
        
        // 添加消息后滚动到底部
        scrollToBottom();
    }

    // 格式化消息内容的函数
    function formatMessageContent(content) {
        if (!content) return '';
        
        // 将普通换行转换为Markdown换行
        content = content.replace(/\n/g, '  \n');
        
        // 使用marked处理Markdown格式
        const markedContent = marked.parse(content);
        
        return markedContent;
    }

    // 修改发送消息函数
    async function sendMessage(content) {
        try {
            if (!content || !currentSessionId) return;

            // 获取复选框的当前状态
            const useContext = document.getElementById('contextCheckbox')?.checked || false;

            // 构建请求数据
            const requestData = {
                sessionId: currentSessionId,
                providerId: parseInt(providerSelect.value),
                modelId: parseInt(modelSelect.value),
                keyId: parseInt(keySelect.value),
                content: content,
                userId: currentUser.id,
                parameters: {
                    enable_internet: internetEnabled,
                    enable_deep_thought: deepThoughtEnabled,
                    use_context: useContext  // 是否使用上下文
                }
            };

            // 清空输入框并重置高度
            messageInput.value = '';
            messageInput.style.height = 'auto';

            // 添加用户消息
            addMessage(content, 'user');

            // 添加AI响应占位
            const aiMessageDiv = document.createElement('div');
            aiMessageDiv.className = 'message assistant';
            aiMessageDiv.innerHTML = `
                <div class="loading-animation">
                    <div class="loading-dots">
                        <span></span>
                        <span></span>
                        <span></span>
                    </div>
                </div>
            `;
            chatMessages.appendChild(aiMessageDiv);

            // 发送请求
            const response = await fetch('/api/chat/messages', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'text/event-stream',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: JSON.stringify(requestData)
            });

            //console.log('收到响应:', response.status, response.headers.get('Content-Type'));
            // 滚动到底部
            scrollToBottom();

            // 检查响应类型
            const contentType = response.headers.get('Content-Type');
            
            if (contentType && contentType.includes('text/event-stream')) {
                //console.log('使用流式传输模式');
                let aiResponse = '';

                // 使用 ReadableStream 而不是 EventSource
                const reader = response.body.getReader();
                const decoder = new TextDecoder();

                while (true) {
                    const {done, value} = await reader.read();
                    if (done) {
                        //console.log('流式传输完成');
                        break;
                    }

                    const chunk = decoder.decode(value);
                    //console.log('收到数据块:', chunk);

                    // 处理数据块
                    const lines = chunk.split('\n');
                    for (const line of lines) {
                        // 跳过空行
                        if (!line.trim()) continue;

                        // 处理事件行
                        if (line.startsWith('event:')) continue;

                        // 处理数据行
                        if (line.startsWith('data:')) {
                            try {
                                const data = JSON.parse(line.slice(5));  // 改为 slice(5) 因为 'data:' 长度为5
                                //console.log('解析后的数据:', data);

                                switch(data.type) {
                                    case 'start':
                                        // 更新用户消息ID
                                        const userMessageDiv = chatMessages.lastElementChild.previousElementSibling;
                                        if (userMessageDiv) {
                                            userMessageDiv.dataset.messageId = data.userMessageId;
                                        }
                                        break;

                                    case 'content':
                                        if (data.content) {
                                            // 追加新内容
                                            aiResponse += data.content;
                                            
                                            // 检查是否包含思维过程和最终回答的格式
                                            let displayContent = aiResponse;
                                            
                                            // 处理 <think> 标签格式
                                            const thinkMatch = aiResponse.match(/<think>([\s\S]*?)<\/think>([\s\S]*)/);
                                            if (thinkMatch) {
                                                const [_, thinkContent, finalAnswer] = thinkMatch;
                                                displayContent = `思维过程：\n${thinkContent.trim()}\n\n最终回答：\n${finalAnswer.trim()}`;
                                            } 
                                            // 处理原有的格式
                                            else if (data.reasoning_content) {
                                                displayContent = `思维过程：\n${data.reasoning_content}\n\n最终回答：\n${data.content}`;
                                            } else {
                                                // 尝试从内容中识别思维过程格式
                                                const parts = aiResponse.split('\n\n');
                                                if (parts.length >= 2 && parts[0].startsWith('思维过程：')) {
                                                    displayContent = aiResponse; // 已经是正确格式，直接使用
                                                }
                                            }
                                            
                                            // 使用模板创建消息
                                            aiMessageDiv.className = 'message assistant';
                                            aiMessageDiv.innerHTML = MESSAGE_TEMPLATE.assistant(marked.parse(displayContent));

                                            // 强制浏览器重绘
                                            void aiMessageDiv.offsetHeight;

                                            // 移除加载动画
                                            const loadingDiv = aiMessageDiv.querySelector('.loading-animation');
                                            if (loadingDiv) {
                                                loadingDiv.remove();
                                            }

                                            // 重新绑定事件监听器
                                            const copyButton = aiMessageDiv.querySelector('.copy-btn');
                                            const deleteButton = aiMessageDiv.querySelector('.delete-btn');
                                            
                                            if (copyButton) {
                                                copyButton.addEventListener('click', () => {
                                                    const finalAnswer = aiMessageDiv.querySelector('.final-answer');
                                                    const content = finalAnswer ? finalAnswer.textContent.replace('最终回答：', '').trim() : aiMessageDiv.querySelector('.message-content').textContent;
                                                    
                                                    const textarea = document.createElement('textarea');
                                                    textarea.value = content;
                                                    textarea.style.position = 'fixed';
                                                    textarea.style.opacity = '0';
                                                    document.body.appendChild(textarea);
                                                    
                                                    try {
                                                        textarea.select();
                                                        document.execCommand('copy');
                                                        copyButton.setAttribute('data-original-title', copyButton.getAttribute('title'));
                                                        copyButton.setAttribute('title', '已复制!');
                                                        setTimeout(() => {
                                                            copyButton.setAttribute('title', copyButton.getAttribute('data-original-title'));
                                                        }, 1000);
                                                    } catch (err) {
                                                        console.error('复制失败:', err);
                                                    } finally {
                                                        document.body.removeChild(textarea);
                                                    }
                                                });
                                            }

                                            if (deleteButton) {
                                                deleteButton.addEventListener('click', async () => {
                                                    if (confirm('确定要删除这条消息吗？')) {
                                                        const messageElement = aiMessageDiv;
                                                        const messageId = messageElement.dataset.messageId;
                                                        
                                                        if (!messageId) {
                                                            alert('无法删除消息：消息ID未找到');
                                                            return;
                                                        }

                                                        try {
                                                            const response = await fetch(`/api/chat/messages/${messageId}?userId=${currentUser.id}`, {
                                                                method: 'DELETE',
                                                                headers: {
                                                                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                                                                }
                                                            });

                                                            if (!response.ok) {
                                                                const error = await response.json();
                                                                throw new Error(error.error || '删除失败');
                                                            }

                                                            // 如果是AI回复，同时检查是否需要删除对应的用户消息
                                                            const prevMessage = messageElement.previousElementSibling;
                                                            if (prevMessage && prevMessage.classList.contains('user')) {
                                                                if (confirm('是否同时删除相关的提问消息？')) {
                                                                    const prevMessageId = prevMessage.dataset.messageId;
                                                                    await fetch(`/api/chat/messages/${prevMessageId}?userId=${currentUser.id}`, {
                                                                        method: 'DELETE',
                                                                        headers: {
                                                                            'Authorization': `Bearer ${localStorage.getItem('token')}`
                                                                        }
                                                                    });
                                                                    prevMessage.remove();
                                                                }
                                                            }
                                                            
                                                            messageElement.remove();
                                                            // 滚动到底部
                                                            //scrollToBottom();
                                                        } catch (error) {
                                                            alert('删除消息失败: ' + error.message);
                                                        }
                                                    }
                                                });
                                            }

                                        }
                                        break;

                                    case 'end':
                                        // 添加详细日志
                                        //console.log('收到 end 事件的原始数据:', data);
                                        //console.log('当前 AI 消息元素:', {
                                        //    element: aiMessageDiv,
                                        //    currentId: aiMessageDiv.dataset.messageId,
                                        //    className: aiMessageDiv.className,
                                        //    innerHTML: aiMessageDiv.innerHTML
                                        //});

                                        // 设置AI消息ID
                                        aiMessageDiv.dataset.messageId = data.assistantMessageId;
                                        // 滚动到底部
                                        scrollToBottom();
                                        
                                        // 验证ID设置结果
                                        //console.log('设置ID后的消息元素:', {
                                        //    newId: aiMessageDiv.dataset.messageId,
                                        //    element: aiMessageDiv,
                                        //    querySelector: document.querySelector(`[data-message-id="${data.assistantMessageId}"]`)
                                        //});

                                        // 检查消息元素的父子关系
                                        //console.log('消息元素的DOM结构:', {
                                        //    parent: aiMessageDiv.parentElement,
                                        //    wrapper: aiMessageDiv.querySelector('.message-wrapper'),
                                        //    hasDeleteBtn: !!aiMessageDiv.querySelector('.delete-btn')
                                        //});
                                        break;

                                    case 'error':
                                        throw new Error(data.error);
                                }
                            } catch (error) {
                                //console.error('处理SSE数据失败:', error, line);
                                continue;
                            }
                        }
                    }
                }

            } else {
                //console.log('使用普通传输模式');
                // 处理普通响应
                const data = await response.json();
                
                if (!response.ok) {
                    throw new Error(data.error || '发送失败');
                }

                // 更新用户消息的ID
                const userMessageDiv = chatMessages.lastElementChild.previousElementSibling;
                if (userMessageDiv) {
                    userMessageDiv.dataset.messageId = data.userMessageId;
                }

                // 使用相同的消息模板更新AI响应
                aiMessageDiv.className = 'message assistant';
                aiMessageDiv.innerHTML = MESSAGE_TEMPLATE.assistant(marked.parse(data.content));
                aiMessageDiv.dataset.messageId = data.assistantMessageId;

                // 移除加载动画
                const loadingDiv = aiMessageDiv.querySelector('.loading-animation');
                if (loadingDiv) {
                    loadingDiv.remove();
                }

                // 绑定事件监听器
                const copyButton = aiMessageDiv.querySelector('.copy-btn');
                const deleteButton = aiMessageDiv.querySelector('.delete-btn');
                
                if (copyButton) {
                    copyButton.addEventListener('click', () => {
                        const finalAnswer = aiMessageDiv.querySelector('.final-answer');
                        const content = finalAnswer ? finalAnswer.textContent.replace('最终回答：', '').trim() : aiMessageDiv.querySelector('.message-content').textContent;
                        
                        const textarea = document.createElement('textarea');
                        textarea.value = content;
                        textarea.style.position = 'fixed';
                        textarea.style.opacity = '0';
                        document.body.appendChild(textarea);
                        
                        try {
                            textarea.select();
                            document.execCommand('copy');
                            copyButton.setAttribute('data-original-title', copyButton.getAttribute('title'));
                            copyButton.setAttribute('title', '已复制!');
                            setTimeout(() => {
                                copyButton.setAttribute('title', copyButton.getAttribute('data-original-title'));
                            }, 1000);
                        } catch (err) {
                            console.error('复制失败:', err);
                        } finally {
                            document.body.removeChild(textarea);
                        }
                    });
                }

                if (deleteButton) {
                    deleteButton.addEventListener('click', async () => {
                        if (confirm('确定要删除这条消息吗？')) {
                            const messageElement = aiMessageDiv;
                            const messageId = messageElement.dataset.messageId;
                            
                            if (!messageId) {
                                alert('无法删除消息：消息ID未找到');
                                return;
                            }

                            try {
                                const response = await fetch(`/api/chat/messages/${messageId}?userId=${currentUser.id}`, {
                                    method: 'DELETE',
                                    headers: {
                                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                                    }
                                });

                                if (!response.ok) {
                                    const error = await response.json();
                                    throw new Error(error.error || '删除失败');
                                }

                                // 如果是AI回复，同时检查是否需要删除对应的用户消息
                                const prevMessage = messageElement.previousElementSibling;
                                if (prevMessage && prevMessage.classList.contains('user')) {
                                    if (confirm('是否同时删除相关的提问消息？')) {
                                        const prevMessageId = prevMessage.dataset.messageId;
                                        await fetch(`/api/chat/messages/${prevMessageId}?userId=${currentUser.id}`, {
                                            method: 'DELETE',
                                            headers: {
                                                'Authorization': `Bearer ${localStorage.getItem('token')}`
                                            }
                                        });
                                        prevMessage.remove();
                                    }
                                }
                                
                                messageElement.remove();
                            } catch (error) {
                                alert('删除消息失败: ' + error.message);
                            }
                        }
                    });
                }

                // 滚动到底部
                scrollToBottom();
            }

        } catch (error) {
            console.error('发送消息失败:', error);
            if (aiMessageDiv) {
                // 移除加载动画
                const loadingDiv = aiMessageDiv.querySelector('.loading-animation');
                if (loadingDiv) {
                    loadingDiv.remove();
                }
                // 显示错误消息
                aiMessageDiv.innerHTML = `<div class="error">发送失败: ${error.message}</div>`;
            }
        }
    }

    // 初始化时加载会话列表和最新会话
    function initialize() {
        console.log('开始加载会话列表');
        loadProviders();
        loadSessions();
        console.log('会话列表加载完成');
        // 添加标签页功能
        const tabButtons = document.querySelectorAll('.tab-btn');
        const tabPanes = document.querySelectorAll('.tab-pane');
        const contextCheckbox = document.getElementById('contextCheckbox');

        // 标签切换功能
        tabButtons.forEach(button => {
            button.addEventListener('click', () => {
                // 移除所有活动状态
                tabButtons.forEach(btn => btn.classList.remove('active'));
                tabPanes.forEach(pane => pane.classList.remove('active'));
                
                // 激活当前标签
                button.classList.add('active');
                const tabId = button.dataset.tab;
                document.getElementById(`${tabId}-pane`).classList.add('active');
            });
        });

        // 上下文设置处理
        let useContext = localStorage.getItem('useContext') === 'true';
        if (contextCheckbox) {
            contextCheckbox.checked = useContext;
            contextCheckbox.addEventListener('change', (e) => {
                useContext = e.target.checked;
                localStorage.setItem('useContext', useContext);
            });
        }
    }

    // 修改加载会话列表函数
    function loadSessions() {
        console.log('开始加载会话列表');
        return fetch(`/api/chat/sessions?userId=${currentUser.id}`,{
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        })
            .then(response => response.json())
            .then(data => {
                const sessionList = document.getElementById('sessionList');
                sessionList.innerHTML = '';
                
                if (!data.items || data.items.length === 0) {
                    sessionList.innerHTML = `
                        <div class="session-empty">
                            <p>暂无会话记录</p>
                            <p>点击"新会话"开始对话</p>
                        </div>
                    `;
                    return;
                }

                data.items.forEach(session => {
                    const sessionDiv = document.createElement('div');
                    sessionDiv.className = 'session-item';
                    sessionDiv.innerHTML = `
                        <div class="session-header">
                            <div class="session-title" title="双击修改标题">
                                ${session.title || '未命名会话'}
                            </div>
                            <button class="session-close" title="删除会话">×</button>
                        </div>
                        <div class="session-info">
                            <span class="session-time">
                                ${formatTime(new Date(session.createdAt))}
                            </span>
                            <span class="message-count">
                                ${session.messageCount || 0}条消息
                            </span>
                        </div>
                    `;

                    // 添加双击编辑标题功能
                    const titleDiv = sessionDiv.querySelector('.session-title');
                    titleDiv.addEventListener('dblclick', (e) => {
                        e.stopPropagation(); // 阻止事件冒泡
                        const input = document.createElement('input');
                        input.type = 'text';
                        input.className = 'session-title-input';
                        input.value = session.title || '未命名会话';
                        titleDiv.replaceWith(input);
                        input.focus();

                        // 处理输入完成
                        const handleTitleUpdate = () => {
                            const newTitle = input.value.trim() || '未命名会话';
                            updateSessionTitle(session.id, newTitle).then(() => {
                                input.replaceWith(titleDiv);
                                titleDiv.textContent = newTitle;
                            });
                        };

                        input.addEventListener('blur', handleTitleUpdate);
                        input.addEventListener('keypress', (e) => {
                            if (e.key === 'Enter') {
                                handleTitleUpdate();
                            }
                        });
                    });

                    // 添加删除按钮事件
                    const closeBtn = sessionDiv.querySelector('.session-close');
                    closeBtn.addEventListener('click', (e) => {
                        e.stopPropagation(); // 阻止事件冒泡
                        if (confirm('确定要删除这个会话吗？此操作不可恢复。')) {
                            deleteSession(session.id);
                        }
                    });

                    // 修改点击会话切换的事件处理
                    sessionDiv.addEventListener('click', () => {
                        // 移除所有会话的激活状态
                        const allSessions = document.querySelectorAll('.session-item');
                        allSessions.forEach(item => item.classList.remove('active'));
                        
                        // 添加当前会话的激活状态
                        sessionDiv.classList.add('active');
                        
                        // 加载会话
                        loadSession(session.id);
                    });

                    sessionList.appendChild(sessionDiv);
                });

                // 如果没有当前会话，自动点击第一个会话
                if (!currentSessionId && data.items.length > 0) {
                    const firstSession = sessionList.querySelector('.session-item');
                    if (firstSession) {
                        firstSession.click();  // 触发点击事件
                    }
                }
            });
    }

    // 格式化时间
    function formatTime(date) {
        const now = new Date();
        const diff = now - date;
        const oneDay = 24 * 60 * 60 * 1000;
        
        if (diff < oneDay) {
            return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
        } else if (diff < 7 * oneDay) {
            const days = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
            return days[date.getDay()];
        } else {
            return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' });
        }
    }

    // 修改加载会话消息函数
    async function loadSessionMessages(sessionId) {
        try {
            console.log('开始加载会话消息');
            const response = await fetch(`/api/chat/sessions/${sessionId}/messages`);
            const data = await response.json();
            
            // 清空现有消息
            chatMessages.innerHTML = '';
            
            // 添加历史消息（包含完整的消息对象，含ID）
            data.messages.forEach(message => {
                addMessage({
                    id: message.id,          // 数据库中的ID
                    content: message.content,
                    role: message.role
                }, message.role);
            });
        } catch (error) {
            //console.error('加载会话消息失败:', error);
        }
    }

    // 修改加载会话消息函数
    function loadSession(sessionId) {
        currentSessionId = sessionId;
        console.log('开始加载会话消息');
        // 先获取会话详情
        fetch(`/api/chat/sessions/${sessionId}?userId=${currentUser.id}`,{
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        })
            .then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            })
            .then(session => {
                // 设置服务商、模型和密钥
                providerSelect.value = session.providerId;
                
                // 加载对应的模型列表和密钥列表
                Promise.all([
                    loadModels(session.providerId).then(() => {
                        modelSelect.value = session.modelId;
                        // 在设置完模型值后检查参数
                        checkModelParameters();
                    }),
                    loadKeys(session.providerId).then(() => {
                        keySelect.value = session.keyId;
                    })
                ]).then(() => {
                    updateSendButton();
                });

                // 加载会话消息
                return fetch(`/api/chat/sessions/${sessionId}/messages?userId=${currentUser.id}`, {
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                    }
                });
            })
            .then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            })
            .then(data => {
                chatMessages.innerHTML = '';
                
                if (Array.isArray(data)) {  // 直接是消息数组
                    data.forEach(msg => {
                        const messageDiv = document.createElement('div');
                        messageDiv.className = `message ${msg.role}`;
                        messageDiv.dataset.messageId = msg.id;
                        
                        // 处理消息内容
                        let content = msg.content;
                        if (content.includes('<think>') && content.includes('</think>')) {
                            const thinkMatch = content.match(/<think>([\s\S]*?)<\/think>([\s\S]*)/);
                            if (thinkMatch) {
                                const [_, thinking, answer] = thinkMatch;
                                content = `思维过程：${thinking.trim()}\n\n最终回答：${answer.trim()}`;
                            }
                        }
                        
                        // 使用 marked 处理 markdown 格式
                        const formattedContent = marked.parse(content);
                        
                        // 为用户消息添加按钮
                        if (msg.role === 'user') {
                            messageDiv.dataset.content = msg.content;  // 保存原始内容用于重发
                            messageDiv.innerHTML = `
                                <div class="message-wrapper">
                                    <div class="message-content markdown-body">${formattedContent}</div>
                                    <div class="message-actions">
                                        <button class="action-btn copy-btn" title="复制消息">
                                            <i class="fas fa-copy"></i>
                                        </button>
                                        <button class="action-btn resend-btn" title="重新发送">
                                            <i class="fas fa-redo-alt"></i>
                                        </button>
                                        <button class="action-btn delete-btn" title="删除消息">
                                            <i class="fas fa-trash"></i>
                                        </button>
                                    </div>
                                </div>
                            `;
                            
                            // 添加按钮事件
                            const copyBtn = messageDiv.querySelector('.copy-btn');
                            const resendBtn = messageDiv.querySelector('.resend-btn');
                            const deleteBtn = messageDiv.querySelector('.delete-btn');

                            // 复制按钮事件
                            copyBtn.addEventListener('click', () => {
                                const textarea = document.createElement('textarea');
                                textarea.value = msg.content;
                                textarea.style.position = 'fixed';
                                textarea.style.opacity = '0';
                                document.body.appendChild(textarea);
                                
                                try {
                                    textarea.select();
                                    document.execCommand('copy');
                                    copyBtn.setAttribute('data-original-title', copyBtn.getAttribute('title'));
                                    copyBtn.setAttribute('title', '已复制!');
                                    setTimeout(() => {
                                        copyBtn.setAttribute('title', copyBtn.getAttribute('data-original-title'));
                                    }, 1000);
                                } catch (err) {
                                    console.error('复制失败:', err);
                                } finally {
                                    document.body.removeChild(textarea);
                                }
                            });

                            // 重发按钮事件
                            resendBtn.addEventListener('click', () => {
                                const nextMessage = messageDiv.nextElementSibling;
                                if (nextMessage && nextMessage.classList.contains('assistant')) {
                                    nextMessage.remove();
                                }
                                sendMessage(msg.content);
                            });

                            // 删除按钮事件
                            deleteBtn.addEventListener('click', async () => {
                                if (confirm('确定要删除这条消息吗？')) {
                                    const messageId = messageDiv.dataset.messageId;
                                    
                                    if (!messageId) {
                                        alert('无法删除消息：消息ID未找到');
                                        return;
                                    }

                                    try {
                                        const response = await fetch(`/api/chat/messages/${messageId}?userId=${currentUser.id}`, {
                                            method: 'DELETE',
                                            headers: {
                                                'Authorization': `Bearer ${localStorage.getItem('token')}`
                                            }
                                        });

                                        if (!response.ok) {
                                            const error = await response.json();
                                            throw new Error(error.error || '删除失败');
                                        }

                                        const nextMessage = messageDiv.nextElementSibling;
                                        if (nextMessage && nextMessage.classList.contains('assistant')) {
                                            const nextMessageId = nextMessage.dataset.messageId;
                                            await fetch(`/api/chat/messages/${nextMessageId}?userId=${currentUser.id}`, {
                                                method: 'DELETE',
                                                headers: {
                                                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                                                }
                                            });
                                            nextMessage.remove();
                                        }
                                        
                                        messageDiv.remove();
                                    } catch (error) {
                                        alert('删除消息失败: ' + error.message);
                                    }
                                }
                            });
                        } else {
                            // AI 消息的处理
                            messageDiv.innerHTML = MESSAGE_TEMPLATE.assistant(formattedContent);
                            
                            // 为 AI 消息添加按钮事件
                            const copyBtn = messageDiv.querySelector('.copy-btn');
                            const deleteBtn = messageDiv.querySelector('.delete-btn');
                            
                            //console.log('加载历史AI消息的按钮:', {copyBtn, deleteBtn}); // 调试日志
                            
                            // 复制按钮事件
                            if (copyBtn) {
                                copyBtn.addEventListener('click', () => {
                                    //console.log('点击复制按钮'); // 调试日志
                                    const finalAnswer = messageDiv.querySelector('.final-answer');
                                    const content = finalAnswer ? 
                                        finalAnswer.textContent.replace('最终回答：', '').trim() : 
                                        messageDiv.querySelector('.message-content').textContent;
                                    
                                    //console.log('要复制的内容:', content); // 调试日志
                                    
                                    const textarea = document.createElement('textarea');
                                    textarea.value = content;
                                    textarea.style.position = 'fixed';
                                    textarea.style.opacity = '0';
                                    document.body.appendChild(textarea);
                                    
                                    try {
                                        textarea.select();
                                        document.execCommand('copy');
                                        copyBtn.setAttribute('data-original-title', copyBtn.getAttribute('title'));
                                        copyBtn.setAttribute('title', '已复制!');
                                        setTimeout(() => {
                                            copyBtn.setAttribute('title', copyBtn.getAttribute('data-original-title'));
                                        }, 1000);
                                        //console.log('复制成功'); // 调试日志
                                    } catch (err) {
                                        console.error('复制失败:', err);
                                    } finally {
                                        document.body.removeChild(textarea);
                                    }
                                });
                            }
                            
                            // 删除按钮事件
                            if (deleteBtn) {
                                deleteBtn.addEventListener('click', async () => {
                                    //console.log('点击删除按钮'); // 调试日志
                                    if (confirm('确定要删除这条消息吗？')) {
                                        const messageId = messageDiv.dataset.messageId;
                                        //console.log('要删除的消息ID:', messageId); // 调试日志
                                        
                                        if (!messageId) {
                                            alert('无法删除消息：消息ID未找到');
                                            return;
                                        }

                                        try {
                                            const response = await fetch(`/api/chat/messages/${messageId}?userId=${currentUser.id}`, {
                                                method: 'DELETE',
                                                headers: {
                                                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                                                }
                                            });

                                            //console.log('删除请求响应:', response); // 调试日志

                                            if (!response.ok) {
                                                const error = await response.json();
                                                throw new Error(error.error || '删除失败');
                                            }

                                            // 如果是AI回复，同时检查是否需要删除对应的用户消息
                                            const prevMessage = messageDiv.previousElementSibling;
                                            if (prevMessage && prevMessage.classList.contains('user')) {
                                                if (confirm('是否同时删除相关的提问消息？')) {
                                                    const prevMessageId = prevMessage.dataset.messageId;
                                                    await fetch(`/api/chat/messages/${prevMessageId}?userId=${currentUser.id}`, {
                                                        method: 'DELETE',
                                                        headers: {
                                                            'Authorization': `Bearer ${localStorage.getItem('token')}`
                                                        }
                                                    });
                                                    prevMessage.remove();
                                                }
                                            }
                                            
                                            messageDiv.remove();
                                            //console.log('删除成功'); // 调试日志
                                        } catch (error) {
                                            //console.error('删除失败:', error);
                                            alert('删除消息失败: ' + error.message);
                                        }
                                    }
                                });
                            }
                        }
                        
                        chatMessages.appendChild(messageDiv);
                    });
                } else if (data.messages && Array.isArray(data.messages)) {  // 包装在 messages 字段中
                    data.messages.forEach(msg => {
                        // 检查消息内容是否包含思维过程
                        let content = msg.content;
                        if (content.includes('<think>') && content.includes('</think>')) {
                            // 处理 <think> 标签格式
                            const thinkMatch = content.match(/<think>([\s\S]*?)<\/think>([\s\S]*)/);
                            if (thinkMatch) {
                                const [_, thinking, answer] = thinkMatch;
                                content = `思维过程：${thinking.trim()}\n\n最终回答：${answer.trim()}`;
                            }
                        }
                        
                        // 使用 marked 处理 markdown 格式
                        const formattedContent = marked.parse(content);
                        
                        // 创建消息元素
                        const messageDiv = document.createElement('div');
                        messageDiv.className = `message ${msg.role}`;
                        messageDiv.dataset.messageId = msg.id;
                        
                        // 使用消息模板
                        if (msg.role === 'assistant') {
                            messageDiv.innerHTML = MESSAGE_TEMPLATE.assistant(formattedContent);
                        } else {
                            messageDiv.innerHTML = `
                                <div class="message-wrapper">
                                    <div class="message-content markdown-body">${formattedContent}</div>
                                </div>
                            `;
                        }
                        
                        chatMessages.appendChild(messageDiv);
                        
                        // 为新添加的消息绑定事件
                        if (msg.role === 'assistant') {
                            const copyBtn = messageDiv.querySelector('.copy-btn');
                            const deleteBtn = messageDiv.querySelector('.delete-btn');
                            
                            //log('AI消息按钮:', {copyBtn, deleteBtn}); // 调试日志
                            
                            // 复制按钮事件
                            if (copyBtn) {
                                copyBtn.addEventListener('click', () => {
                                    //console.log('点击复制按钮'); // 调试日志
                                    const finalAnswer = messageDiv.querySelector('.final-answer');
                                    const content = finalAnswer ? 
                                        finalAnswer.textContent.replace('最终回答：', '').trim() : 
                                        messageDiv.querySelector('.message-content').textContent;
                                    
                                    //console.log('要复制的内容:', content); // 调试日志
                                    
                                    const textarea = document.createElement('textarea');
                                    textarea.value = content;
                                    textarea.style.position = 'fixed';
                                    textarea.style.opacity = '0';
                                    document.body.appendChild(textarea);
                                    
                                    try {
                                        textarea.select();
                                        document.execCommand('copy');
                                        copyBtn.setAttribute('data-original-title', copyBtn.getAttribute('title'));
                                        copyBtn.setAttribute('title', '已复制!');
                                        setTimeout(() => {
                                            copyBtn.setAttribute('title', copyBtn.getAttribute('data-original-title'));
                                        }, 1000);
                                        //console.log('复制成功'); // 调试日志
                                    } catch (err) {
                                        //console.error('复制失败:', err);
                                    } finally {
                                        document.body.removeChild(textarea);
                                    }
                                });
                            }
                            
                            // 删除按钮事件
                            if (deleteBtn) {
                                deleteBtn.addEventListener('click', async () => {
                                    //console.log('点击删除按钮'); // 调试日志
                                    if (confirm('确定要删除这条消息吗？')) {
                                        const messageId = messageDiv.dataset.messageId;
                                        //console.log('要删除的消息ID:', messageId); // 调试日志
                                        
                                        if (!messageId) {
                                            alert('无法删除消息：消息ID未找到');
                                            return;
                                        }

                                        try {
                                            const response = await fetch(`/api/chat/messages/${messageId}?userId=${currentUser.id}`, {
                                                method: 'DELETE',
                                                headers: {
                                                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                                                }
                                            });

                                            //console.log('删除请求响应:', response); // 调试日志

                                            if (!response.ok) {
                                                const error = await response.json();
                                                throw new Error(error.error || '删除失败');
                                            }

                                            // 如果是AI回复，同时检查是否需要删除对应的用户消息
                                            const prevMessage = messageDiv.previousElementSibling;
                                            if (prevMessage && prevMessage.classList.contains('user')) {
                                                if (confirm('是否同时删除相关的提问消息？')) {
                                                    const prevMessageId = prevMessage.dataset.messageId;
                                                    await fetch(`/api/chat/messages/${prevMessageId}?userId=${currentUser.id}`, {
                                                        method: 'DELETE',
                                                        headers: {
                                                            'Authorization': `Bearer ${localStorage.getItem('token')}`
                                                        }
                                                    });
                                                    prevMessage.remove();
                                                }
                                            }
                                            
                                            messageDiv.remove();
                                            //console.log('删除成功'); // 调试日志
                                        } catch (error) {
                                            //console.error('删除失败:', error);
                                            alert('删除消息失败: ' + error.message);
                                        }
                                    }
                                });
                            }
                        }
                    });
                }
                
                // 启用输入框
                messageInput.disabled = false;
                updateSendButton();
                
                // 滚动到底部
                scrollToBottom();
            })
            .catch(error => {
                //console.error('加载会话失败:', error);
                chatMessages.innerHTML = `
                    <div class="message error">
                        <div class="message-content">加载会话失败: ${error.message}</div>
                    </div>
                `;
            });
    }

    // 修改创建新会话函数
    function createNewSession() {
        // 获取当前选中的值
        const selectedProviderId = parseInt(providerSelect.value) || 13;  // 如果没有选择，使用默认值
        const selectedModelId = parseInt(modelSelect.value) || 10;
        const selectedKeyId = parseInt(keySelect.value) || 15;

        // 获取选中的模型名称
        const selectedModelName = modelSelect.options[modelSelect.selectedIndex]?.text || '新会话';

        fetch('/api/chat/sessions', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify({
                userId: currentUser.id,
                title: selectedModelName,  // 使用模型名称作为会话标题
                providerId: selectedProviderId,
                modelId: selectedModelId,
                keyId: selectedKeyId
            })
        })
        .then(response => response.json())
        .then(data => {
            loadSessions().then(() => {
                // 自动点击新创建的会话
                const newSession = document.querySelector('.session-item');
                if (newSession) {
                    newSession.click();  // 触发点击事件
                }
            });
        });
    }

    // 修改更新发送按钮状态函数
    function updateSendButton() {
        const hasValidSession = currentSessionId != null;
        const hasValidProvider = providerSelect.value !== "";
        const hasValidModel = modelSelect.value !== "";
        const hasValidKey = keySelect.value !== "";
        
        const isValid = hasValidSession && hasValidProvider && hasValidModel && hasValidKey;
        sendButton.disabled = !isValid;
        
        // 添加调试日志
        //console.log('发送按钮状态检查:', {
        //    hasValidSession,
        //    hasValidProvider,
        //    hasValidModel,
        //    hasValidKey,
        //    isValid
        //});
    }

    // 事件监听
    providerSelect.addEventListener('change', function() {
        const providerId = this.value;
        loadModels(providerId);
        loadKeys(providerId);
        updateSendButton();
    });

    // 添加模型选择事件监听
    modelSelect.addEventListener('change', function() {
        //console.log('模型选择改变');
        checkModelParameters();
        updateSendButton();
    });

    keySelect.addEventListener('change', updateSendButton);

    // 修改输入框事件监听
    messageInput.addEventListener('keydown', (e) => {
        // Shift + Enter 换行
        if (e.key === 'Enter' && e.shiftKey) {
            return; // 允许换行
        }
        
        // 仅 Enter 发送消息
        if (e.key === 'Enter') {
            e.preventDefault();
            const content = messageInput.value.trim();
            if (content) {
                sendMessage(content);
            }
        }
    });

    // 自动调整输入框高度
    messageInput.addEventListener('input', function() {
        // 重置高度
        this.style.height = 'auto';
        
        // 设置新高度
        const newHeight = Math.min(this.scrollHeight, 200); // 最大高度 200px
        this.style.height = newHeight + 'px';
    });

    // 修改发送按钮点击事件
    sendButton.addEventListener('click', () => {
        const content = messageInput.value.trim();
        if (content) {
            sendMessage(content);
        }
    });

    // 添加更新会话标题的函数
    function updateSessionTitle(sessionId, newTitle) {
        return fetch(`/api/chat/sessions/${sessionId}?userId=${currentUser.id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify({ title: newTitle })
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('更新标题失败');
            }
            return response.json();
        })
        .catch(error => {
            //console.error('更新会话标题失败:', error);
            alert('更新会话标题失败');
        });
    }

    // 添加删除会话的函数
    function deleteSession(sessionId) {
        fetch(`/api/chat/sessions/${sessionId}?userId=${currentUser.id}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('删除会话失败');
            }
            // 如果删除的是当前会话，清空消息区域
            if (sessionId === currentSessionId) {
                currentSessionId = null;
                chatMessages.innerHTML = '';
                messageInput.disabled = true;
                updateSendButton();
            }
            // 重新加载会话列表
            loadSessions();
        })
        .catch(error => {
            //console.error('删除会话失败:', error);
            alert('删除会话失败');
        });
    }

    // 绑定新会话按钮点击事件
    newSessionBtn.addEventListener('click', () => {
        // 移除旧的事件监听器，使用新的 createNewSession 函数
        createNewSession();
    });

    // 添加滚动到底部的函数
    function scrollToBottom() {
        const chatMessages = document.getElementById('chatMessages');
        chatMessages.scrollTop = chatMessages.scrollHeight;
        //console.log('滚动到底部');
    }

    // 启动初始化
    initialize();
}); 
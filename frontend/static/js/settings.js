// 系统设置页面的 JavaScript
document.addEventListener('DOMContentLoaded', function() {
    // 初始化设置页面
    async function initializeSettings() {
        try {
            // 获取所有标签按钮和面板
            const tabButtons = document.querySelectorAll('.settings-tabs .tab-btn');
            const tabPanels = document.querySelectorAll('.tab-panel');
            
            // 标签切换功能
            tabButtons.forEach(button => {
                button.addEventListener('click', () => {
                    // 移除所有活动状态
                    tabButtons.forEach(btn => btn.classList.remove('active'));
                    tabPanels.forEach(panel => panel.classList.remove('active'));
                    
                    // 激活当前标签
                    button.classList.add('active');
                    const tabId = button.dataset.tab;
                    document.getElementById(`${tabId}-panel`).classList.add('active');
                });
            });

            // Redis设置相关
            const redisForm = document.getElementById('redisSettingsForm');
            const testRedisBtn = document.getElementById('testRedisBtn');

            // 加载Redis设置
            loadRedisSettings();

            // 测试Redis连接
            testRedisBtn?.addEventListener('click', testRedisConnection);

            // 保存Redis设置
            redisForm?.addEventListener('submit', saveRedisSettings);

            // 搜索设置相关
            const searchForm = document.getElementById('searchSettingsForm');
            
            // 加载搜索设置
            loadSearchSettings();

            // 保存搜索设置
            searchForm?.addEventListener('submit', saveSearchSettings);

            // 对话设置相关
            const chatForm = document.getElementById('chatSettingsForm');
            
            // 加载对话设置
            loadChatSettings();

            // 保存对话设置
            chatForm?.addEventListener('submit', saveChatSettings);
        } catch (error) {
            console.error('初始化设置失败:', error);
            showError('初始化设置失败');
        }
    }

    // 加载对话设置
    async function loadChatSettings() {
        try {
            const response = await fetch('/api/settings/chat');
            const data = await response.json();
            console.log('加载对话设置:', data);
            // AI请求控制
            document.getElementById('aiRequestTimeout').value = data['ai.request_timeout'] || '';
            
            // 会话控制
            document.getElementById('sessionContextLength').value = data['session.context_length'] || '';
        } catch (error) {
            console.error('加载对话设置失败:', error);
            showError('加载对话设置失败');
        }
    }

    // 保存对话设置
    async function saveChatSettings(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const settings = {};
        
        // 验证并转换数值
        for (const [key, value] of formData.entries()) {
            // 确保数值字段不为空
            if (value === '') {
                showError(`${key} 不能为空`);
                return;
            }
            // 对于数值类型的设置，确保是有效的数字
            if (key === 'ai.request_timeout' || key === 'session.context_length') {
                const num = parseInt(value);
                if (isNaN(num)) {
                    showError(`${key} 必须是数字`);
                    return;
                }
                settings[key] = num.toString();
            } else {
                settings[key] = value;
            }
        }

        try {
            const response = await fetch('/api/settings/chat', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(settings)
            });

            if (!response.ok) {
                throw new Error('保存失败');
            }
            
            showSuccess('对话设置保存成功');
        } catch (error) {
            console.error('保存对话设置失败:', error);
            showError('保存对话设置失败');
        }
    }

    // 加载Redis设置
    async function loadRedisSettings() {
        try {
            const response = await fetch('/api/settings/redis');
            const data = await response.json();
            
            // 填充表单
            document.getElementById('redisHost').value = data['redis.host'] || '';
            document.getElementById('redisPort').value = data['redis.port'] || '';
            document.getElementById('redisPassword').value = data['redis.password'] || '';
            document.getElementById('redisDB').value = data['redis.db'] || '';
        } catch (error) {
            console.error('加载Redis设置失败:', error);
            alert('加载Redis设置失败');
        }
    }

    // 测试Redis连接
    async function testRedisConnection() {
        const config = {
            host: document.getElementById('redisHost').value,
            port: parseInt(document.getElementById('redisPort').value),
            password: document.getElementById('redisPassword').value,
            db: parseInt(document.getElementById('redisDB').value)
        };

        try {
            const response = await fetch('/api/settings/redis/test', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(config)
            });

            const result = await response.json();
            if (result.success) {
                alert('Redis连接测试成功！');
            } else {
                alert('Redis连接测试失败: ' + result.error);
            }
        } catch (error) {
            console.error('测试Redis连接失败:', error);
            alert('测试Redis连接失败');
        }
    }

    // 保存Redis设置
    async function saveRedisSettings(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const settings = Object.fromEntries(formData.entries());

        try {
            const response = await fetch('/api/settings/redis', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(settings)
            });

            if (response.ok) {
                alert('Redis设置保存成功！');
            } else {
                throw new Error('保存失败');
            }
        } catch (error) {
            console.error('保存Redis设置失败:', error);
            alert('保存Redis设置失败');
        }
    }

    // 加载搜索设置
    async function loadSearchSettings() {
        try {
            const response = await fetch('/api/settings/search');
            const data = await response.json();
            
            // 基础设置
            document.getElementById('searchEngineURL').value = data['search.engine_url'];
            document.getElementById('searchResultCount').value = data['search.result_count'];
            document.getElementById('searchMaxKeywords').value = data['search.max_keywords'];
            document.getElementById('searchConcurrency').value = data['search.concurrency'];

            // 性能设置
            document.getElementById('searchTimeout').value = data['search.timeout'];
            document.getElementById('searchRetryCount').value = data['search.retry_count'];
            document.getElementById('searchCacheDuration').value = data['search.cache_duration'];

            // 权重设置
            document.getElementById('weightOfficial').value = data['search.weight.official'];
            document.getElementById('weightEdu').value = data['search.weight.edu'];
            document.getElementById('weightNews').value = data['search.weight.news'];
            document.getElementById('weightPortal').value = data['search.weight.portal'];

            // 时效性权重
            document.getElementById('timeWeightDay').value = data['search.time_weight.day'];
            document.getElementById('timeWeightWeek').value = data['search.time_weight.week'];
            document.getElementById('timeWeightMonth').value = data['search.time_weight.month'];
            document.getElementById('timeWeightYear').value = data['search.time_weight.year'];

            // 过滤设置
            document.getElementById('minContentLength').value = data['search.min_content_length'];
            document.getElementById('maxSummaryLength').value = data['search.max_summary_length'];
            document.getElementById('filterDomains').value = data['search.filter_domains'];

            // 搜索质量控制
            document.getElementById('minTitleLength').value = data['search.min_title_length'];
            document.getElementById('maxResults').value = data['search.max_results'];
            document.getElementById('minRelevanceScore').value = data['search.min_relevance_score'];

            // 缓存控制
            document.getElementById('cacheCleanupInterval').value = data['search.cache_cleanup_interval'];
            document.getElementById('maxCacheSize').value = data['search.max_cache_size'];

            // 请求控制
            document.getElementById('rateLimit').value = data['search.rate_limit'];
            document.getElementById('maxRetryInterval').value = data['search.max_retry_interval'];
            document.getElementById('connectionTimeout').value = data['search.connection_timeout'];

            // 内容过滤
            document.getElementById('blockedKeywords').value = data['search.blocked_keywords'];
            document.getElementById('requiredKeywords').value = data['search.required_keywords'];
            document.getElementById('languageFilter').value = data['search.language_filter'];

        } catch (error) {
            console.error('加载搜索设置失败:', error);
            alert('加载搜索设置失败');
        }
    }

    // 保存搜索设置
    async function saveSearchSettings(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const settings = {};
        
        // 验证并转换数值
        for (const [key, value] of formData.entries()) {
            // 检查是否是数值类型的设置
            if (key.match(/\.(timeout|count|size|length|limit|interval|score|weight|results|concurrency)$/)) {
                // 数值字段必须有值且为有效数字
                if (value === '') {
                    showError(`${key} 不能为空`);
                    return;
                }
                const num = parseInt(value);
                if (isNaN(num)) {
                    showError(`${key} 必须是数字`);
                    return;
                }
                settings[key] = num.toString();
            } else {
                // 非数值字段直接保存，允许为空
                settings[key] = value;
            }
        }

        try {
            const response = await fetch('/api/settings/search', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(settings)
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || '保存失败');
            }
            
            showSuccess('搜索设置保存成功');
        } catch (error) {
            console.error('保存搜索设置失败:', error);
            showError(error.message);
        }
    }

    // 显示成功消息
    function showSuccess(message) {
        alert(message); // 暂时使用 alert，后续可以改为更友好的提示
    }

    // 显示错误消息
    function showError(message) {
        alert(message); // 暂时使用 alert，后续可以改为更友好的提示
    }

    // 启动初始化
    initializeSettings();

    // 统一使用 alert 替换其他地方的错误提示
    function updateAlerts() {
        // 搜索设置相关
        const oldSaveSearchSettings = saveSearchSettings;
        saveSearchSettings = async function(e) {
            try {
                await oldSaveSearchSettings.call(this, e);
            } catch (error) {
                showError(error.message);
            }
        };

        // Redis设置相关
        const oldSaveRedisSettings = saveRedisSettings;
        saveRedisSettings = async function(e) {
            try {
                await oldSaveRedisSettings.call(this, e);
            } catch (error) {
                showError(error.message);
            }
        };
    }
}); 
let currentStep = 1;
const totalSteps = 4;

// 更新进度条
function updateProgress() {
    const progress = (currentStep - 1) / totalSteps * 100;
    document.getElementById('progressBar').style.width = `${progress}%`;
}

// 显示指定步骤
function showStep(step) {
    for (let i = 1; i <= totalSteps; i++) {
        const stepElement = document.getElementById(`step${i}`);
        stepElement.style.display = i === step ? 'block' : 'none';
    }
    currentStep = step;
    updateProgress();
}

// 上一步
function prevStep() {
    if (currentStep > 1) {
        showStep(currentStep - 1);
    }
}

// 下一步
function nextStep() {
    if (currentStep < totalSteps) {
        showStep(currentStep + 1);
    }
}

// 环境检测
async function checkEnvironment() {
    // 更新所有检查项状态为"检测中..."
    const checkItems = ['osCheck', 'cpuCheck', 'memoryCheck', 'diskCheck'];
    checkItems.forEach(id => {
        document.getElementById(id).className = 'check-status status-checking';
        document.getElementById(id).textContent = '检测中...';
    });

    try {
        const response = await axios.get('/api/install/check-environment');
        console.log('Response:', response.data);
        
        const data = response.data;

        if (data && Array.isArray(data.items)) {
            // 更新每个检查项的状态
            const checkMap = {
                'os': 'osCheck',
                'cpu': 'cpuCheck',
                'memory': 'memoryCheck',
                'disk': 'diskCheck'
            };

            data.items.forEach(check => {
                const elementId = checkMap[check.Name];
                const element = document.getElementById(elementId);
                if (element) {
                    element.className = 'check-status ' + (check.Status ? 'status-success' : 'status-error');
                    element.textContent = check.Message;
                }
            });

            // 如果所有检查都通过，允许进入下一步
            if (data.success) {
                setTimeout(() => {
                    nextStep();
                }, 1000);
            } else if (data.errors && data.errors.length > 0) {
                // 显示错误信息
                alert('环境检测失败：\n' + data.errors.join('\n'));
            }
        } else {
            console.error('Invalid response format:', data);
            throw new Error('Invalid response format');
        }
    } catch (error) {
        console.error('Error:', error);
        
        // 显示错误信息
        alert('检测过程出错：' + (error.response?.data?.error || error.message));
        
        // 更新所有检查项状态为错误
        checkItems.forEach(id => {
            const element = document.getElementById(id);
            element.className = 'check-status status-error';
            element.textContent = '检测失败';
        });
    }
}

// 测试数据库连接
async function testDatabase() {
    const dbConfig = {
        type: document.getElementById('dbType').value,
        host: document.getElementById('dbHost').value,
        port: parseInt(document.getElementById('dbPort').value),
        database: document.getElementById('dbName').value,
        username: document.getElementById('dbUser').value,
        password: document.getElementById('dbPassword').value
    };

    try {
        const response = await axios.post('/api/install/test-database', dbConfig);
        if (response.data.success) {
            nextStep();
        } else {
            alert('数据库连接失败：' + response.data.error);
        }
    } catch (error) {
        alert('测试过程出错：' + error.message);
    }
}

// 测试Redis连接
async function testRedis() {
    const redisConfig = {
        host: document.getElementById('redisHost').value,
        port: parseInt(document.getElementById('redisPort').value),
        password: document.getElementById('redisPassword').value,
        db: parseInt(document.getElementById('redisDB').value)
    };

    try {
        const response = await axios.post('/api/install/test-redis', redisConfig);
        if (response.data.success) {
            nextStep();
        } else {
            alert('Redis连接失败：' + response.data.error);
        }
    } catch (error) {
        alert('测试过程出错：' + error.message);
    }
}

// 完成安装
async function completeInstall() {
    // 获取管理员信息
    const adminUsername = document.getElementById('adminUsername').value;
    const adminName = document.getElementById('adminName').value;
    const adminPassword = document.getElementById('adminPassword').value;
    const adminPasswordConfirm = document.getElementById('adminPasswordConfirm').value;
    const adminEmail = document.getElementById('adminEmail').value;

    // 验证表单
    if (!adminUsername) {
        showError('请输入管理员用户名');
        return;
    }
    if (!adminName) {
        showError('请输入管理员姓名');
        return;
    }
    if (!adminPassword) {
        showError('请输入管理员密码');
        return;
    }
    if (adminPassword !== adminPasswordConfirm) {
        showError('两次输入的密码不一致');
        return;
    }
    if (!adminEmail) {
        showError('请输入管理员邮箱');
        return;
    }

    // 发送请求
    fetch('/api/install/complete', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            admin: {
                username: adminUsername,
                name: adminName,
                password: adminPassword,
                email: adminEmail
            }
        })
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            // 显示成功消息
            showSuccess('安装完成！正在跳转到登录页面...');
            
            // 确保显示消息后再跳转
            setTimeout(() => {
                window.location.href = '/login';
            }, 1500); // 调整为1.5秒，给用户足够时间看到成功消息
        } else {
            showError(data.error || '安装失败');
        }
    })
    .catch(error => {
        showError('请求失败: ' + error);
    });
}

// 添加显示成功消息的函数
function showSuccess(message) {
    // 如果页面上没有消息显示区域，创建一个
    let messageDiv = document.getElementById('messageDiv');
    if (!messageDiv) {
        messageDiv = document.createElement('div');
        messageDiv.id = 'messageDiv';
        messageDiv.style.position = 'fixed';
        messageDiv.style.top = '20px';
        messageDiv.style.left = '50%';
        messageDiv.style.transform = 'translateX(-50%)';
        messageDiv.style.padding = '15px 30px';
        messageDiv.style.borderRadius = '5px';
        messageDiv.style.zIndex = '1000';
        document.body.appendChild(messageDiv);
    }

    // 设置成功消息样式
    messageDiv.style.backgroundColor = '#4CAF50';
    messageDiv.style.color = 'white';
    messageDiv.textContent = message;

    // 3秒后自动隐藏消息
    setTimeout(() => {
        messageDiv.style.display = 'none';
    }, 3000);
}

// 初始化显示第一步
showStep(1); 
let currentCaptchaId = '';

// 获取新验证码
async function refreshCaptcha() {
    try {
        const response = await fetch('/api/captcha/generate');
        const data = await response.json();
        
        //console.log('验证码响应:', data);
        if (data.imageBase64) {
            currentCaptchaId = data.captchaId;
            //console.log('设置验证码ID:', currentCaptchaId);
            // 检查 imageBase64 是否已经包含 data:image 前缀
            const imageData = data.imageBase64.startsWith('data:image') 
                ? data.imageBase64 
                : "data:image/png;base64," + data.imageBase64;
            document.getElementById('captchaImg').src = imageData;
        }
    } catch (error) {
        console.error('获取验证码失败:', error);
    }
}

// 初始化函数
function init() {
    //console.log('DOM 加载完成');

    // 确保所有元素都存在
    const refreshButton = document.getElementById('refreshCaptcha');
    if (!refreshButton) {
        console.error('找不到刷新按钮元素');
        return;
    }

    // 页面加载时获取验证码
    refreshCaptcha();

    // 刷新按钮点击事件
    refreshButton.addEventListener('click', (e) => {
        e.preventDefault();
        refreshCaptcha();
    });

    // 登录按钮点击事件
    const loginButton = document.getElementById('loginButton');
    if (loginButton) {
        loginButton.addEventListener('click', login);
    }

    // 按回车键登录
    document.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            login();
        }
    });
}

// 等待 DOM 加载完成后执行初始化
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}

function login() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const captchaCode = document.getElementById('captcha').value;

    // 添加调试日志
    //console.log('登录请求参数:', {
    //    username,
    //    password,
    //    captchaId: currentCaptchaId,
    //    captchaCode
    //});

    // 验证码ID检查
    if (!currentCaptchaId) {
        showMessage('验证码未加载，请刷新页面重试', 'error');
        return;
    }

    if (!username || !password || !captchaCode) {
        showMessage('请输入用户名、密码和验证码', 'error');
        return;
    }

    fetch('/api/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            username,
            password,
            captchaId: currentCaptchaId,
            captchaCode
        })
    })
    .then(response => {
        // 添加调试日志
        //console.log('响应状态:', response.status);
        return response.json();
    })
    .then(data => {
        // 添加调试日志
        //console.log('响应数据:', data);
        if (data.error) {
            showMessage(data.error, 'error');
            refreshCaptcha();
        } else {
            showMessage('登录成功，正在跳转...', 'success');
            localStorage.setItem('user', JSON.stringify(data.user));
            localStorage.setItem('token', data.token);
            
            setTimeout(() => {
                window.location.href = '/home';
            }, 1000);
        }
    })
    .catch(error => {
        //console.error('登录失败:', error);
        showMessage('登录失败，请重试', 'error');
        refreshCaptcha();
    });
}

function showMessage(message, type) {
    const messageDiv = document.querySelector('.message') || createMessageDiv();
    messageDiv.textContent = message;
    messageDiv.className = `message ${type}`;
    messageDiv.style.display = 'block';

    setTimeout(() => {
        messageDiv.style.display = 'none';
    }, 3000);
}

function createMessageDiv() {
    const div = document.createElement('div');
    div.className = 'message';
    document.body.appendChild(div);
    return div;
} 
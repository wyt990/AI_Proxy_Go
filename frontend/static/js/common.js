window.appReady = false; 

document.addEventListener('DOMContentLoaded', function() {
    //console.log('common.js 已加载');
    // 检查登录状态
    const user = JSON.parse(localStorage.getItem('user') || '{}');
    const token = localStorage.getItem('token');

    if (!token) {
        console.log('未登录，重定向到登录页0x001');
        window.location.href = '/login';
        return;
    }

    // 简单验证token (JWT格式验证)
    try {
        const parts = token.split('.');
        if (parts.length !== 3) {
            throw new Error('Invalid token format');
        }
        
        // 检查过期时间
        const payload = JSON.parse(atob(parts[1]));
        if (payload.exp && payload.exp < Math.floor(Date.now() / 1000)) {
            throw new Error('Token expired');
        }
    } catch (e) {
        console.error('Token validation failed:', e);
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        console.log('未登录，重定向到登录页0x002');
        window.location.href = '/login';
        return;
    }

    // 显示用户信息
    const userNameElement = document.getElementById('userName');
    if (userNameElement) {
        userNameElement.textContent = user.name || user.username;
    }

    // 获取所有导航项
    const navItems = document.querySelectorAll('.nav-item');
    //console.log('找到链接数量:', navItems.length);
    
    // 获取搜索框
    const searchInput = document.getElementById('searchInput');
    
    // 更新搜索框提示文本
    function updateSearchPlaceholder(page) {
        if (searchInput) {
            switch(page) {
                case 'dashboard':
                    searchInput.placeholder = '搜索统计数据...';
                    break;
                case 'providers':
                    searchInput.placeholder = '搜索服务商...';
                    break;
                case 'AI_models':
                    searchInput.placeholder = '搜索模型...';
                    break;
                case 'users':
                    searchInput.placeholder = '搜索用户...';
                    break;
                case 'AI_keys':
                    searchInput.placeholder = '搜索密钥...';
                    break;
                case 'chat':
                    searchInput.placeholder = '搜索对话...';
                    break;
                case 'logs':
                    searchInput.placeholder = '搜索日志...';
                    break;
                default:
                    searchInput.placeholder = '搜索...';
            }
        }
    }
    
    // 为每个导航项添加点击事件
    navItems.forEach(item => {
        item.addEventListener('click', function(e) {
            console.log('Nav item clicked:', this.getAttribute('data-page'));
            e.preventDefault();
            
            // 移除所有导航项的active类
            navItems.forEach(nav => nav.classList.remove('active'));
            
            // 为当前点击的导航项添加active类
            this.classList.add('active');
            
            // 获取页面标识
            const page = this.getAttribute('data-page');
            
            // 更新搜索框提示文本
            updateSearchPlaceholder(page);
            
            // 根据页面标识加载不同的内容
            switch(page) {
                case 'dashboard':
                    //console.log('Navigating to /home');
                    window.location.href = '/home';
                    break;
                case 'providers':
                    window.location.href = '/providers';
                    break;
                case 'AI_models':
                    window.location.href = '/AI_models';
                    break;
                case 'users':
                    //console.log('Navigating to /users');
                    window.location.href = '/users';
                    break;
                case 'AI_keys':
                    //console.log('Navigating to /AI_keys');
                    window.location.href = '/AI_keys';
                    break;
                case 'chat':
                    //console.log('Navigating to /chat');
                    window.location.href = `/chat?t=${new Date().getTime()}`;
                    break;
                case 'logs':
                    window.location.href = '/logs';
                    break;
                case 'settings':
                    window.location.href = '/settings';
                    break;
            }
        });
    });

    // 根据当前URL高亮对应的导航项
    const currentPath = window.location.pathname;
    navItems.forEach(item => {
        const page = item.getAttribute('data-page');
        if ((page === 'dashboard' && currentPath === '/home') ||
            (page === 'providers' && currentPath === '/providers') ||
            (page === 'AI_models' && currentPath === '/AI_models') ||
            (page === 'users' && currentPath === '/users') ||
            (page === 'AI_keys' && currentPath === '/AI_keys') ||
            (page === 'logs' && currentPath === '/logs') ||
            (page === 'settings' && currentPath === '/settings') ||
            (page === 'chat' && currentPath === '/chat')) {
            item.classList.add('active');
            updateSearchPlaceholder(page);
        }
    });

    window.appReady = true;
    document.dispatchEvent(new Event('appready'));
});

function logout() {
    fetch('/api/logout', {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            // 清除本地存储
            localStorage.removeItem('user');
            localStorage.removeItem('token');
            
            // 跳转到登录页
            console.log('退出登录，重定向到登录页0x003');
            window.location.href = '/login';
        }
    })
    .catch(error => {
        //console.error('Logout failed:', error);
        // 即使请求失败，也清除本地存储并跳转
        localStorage.removeItem('user');
        localStorage.removeItem('token');
        console.log('退出登录，重定向到登录页0x004');
        window.location.href = '/login';
    });
}

// 检查和同步cookie与localStorage的token
function syncTokenState() {
    const cookieToken = getCookie('token');
    const localToken = localStorage.getItem('token');
    
    // cookie优先级更高，因为服务器可能设置了新的cookie
    if (cookieToken && cookieToken !== localToken) {
        localStorage.setItem('token', cookieToken);
        return true;
    }
    
    // 如果localStorage有token但cookie没有，可能是cookie过期了
    if (!cookieToken && localToken) {
        // 尝试重新请求与服务器验证
        fetch('/api/user/info', {
            headers: {'Authorization': `Bearer ${localToken}`}
        })
        .then(response => {
            if (!response.ok) {
                localStorage.removeItem('token');
                localStorage.removeItem('user');
                console.log('检查和同步cookie与localStorage的token，重定向到登录页0x005');
                window.location.href = '/login';
            }
        });
    }
    
    return cookieToken || localToken;
}

// 获取cookie的辅助函数
function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
}

// 添加全局Ajax请求拦截
function setupAjaxInterceptor() {
    const originalFetch = window.fetch;
    window.fetch = function(url, options = {}) {
        console.log('Fetch intercepted:', url);
        return originalFetch(url, options).then(response => {
            if (response.status === 401) {
                console.log('收到401响应，但不跳转');
                // 不做自动跳转，而是在页面显示提示
                document.body.innerHTML += '<div class="auth-error">登录已过期，<a href="/login">点击重新登录</a></div>';
            }
            return response;
        });
    };
}

// 初始化时设置拦截器
setupAjaxInterceptor(); 
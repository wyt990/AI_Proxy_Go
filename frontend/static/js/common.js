document.addEventListener('DOMContentLoaded', function() {
    //console.log('common.js 已加载');
    // 检查登录状态
    const user = JSON.parse(localStorage.getItem('user') || '{}');
    const token = localStorage.getItem('token');

    if (!token) {
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
                    window.location.href = '/chat';
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
            window.location.href = '/login';
        }
    })
    .catch(error => {
        //console.error('Logout failed:', error);
        // 即使请求失败，也清除本地存储并跳转
        localStorage.removeItem('user');
        localStorage.removeItem('token');
        window.location.href = '/login';
    });
} 
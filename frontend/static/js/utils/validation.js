// 表单验证工具
const validation = {
    // 验证邮箱
    isEmail: function(email) {
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
    },

    // 验证URL
    isURL: function(url) {
        try {
            new URL(url);
            return true;
        } catch {
            return false;
        }
    },

    // 验证密码强度
    checkPasswordStrength: function(password) {
        let strength = 0;
        if (password.length >= 8) strength++;
        if (/[A-Z]/.test(password)) strength++;
        if (/[a-z]/.test(password)) strength++;
        if (/[0-9]/.test(password)) strength++;
        if (/[^A-Za-z0-9]/.test(password)) strength++;
        return strength;
    },

    // 验证手机号
    isMobile: function(mobile) {
        return /^1[3-9]\d{9}$/.test(mobile);
    }
};

// 导出验证工具
window.validation = validation; 
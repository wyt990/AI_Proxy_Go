// 动画工具
const animation = {
    // 淡入效果
    fadeIn: function(element, duration = 300) {
        element.style.opacity = 0;
        element.style.display = 'block';

        let start = null;
        function animate(timestamp) {
            if (!start) start = timestamp;
            const progress = timestamp - start;
            element.style.opacity = Math.min(progress / duration, 1);

            if (progress < duration) {
                window.requestAnimationFrame(animate);
            }
        }
        window.requestAnimationFrame(animate);
    },

    // 淡出效果
    fadeOut: function(element, duration = 300) {
        let start = null;
        const initialOpacity = parseFloat(getComputedStyle(element).opacity);

        function animate(timestamp) {
            if (!start) start = timestamp;
            const progress = timestamp - start;
            element.style.opacity = Math.max(initialOpacity - (progress / duration), 0);

            if (progress < duration) {
                window.requestAnimationFrame(animate);
            } else {
                element.style.display = 'none';
            }
        }
        window.requestAnimationFrame(animate);
    },

    // 滑动效果
    slideDown: function(element, duration = 300) {
        element.style.height = 'auto';
        const height = element.offsetHeight;
        element.style.height = '0px';
        element.style.overflow = 'hidden';
        element.style.display = 'block';

        let start = null;
        function animate(timestamp) {
            if (!start) start = timestamp;
            const progress = timestamp - start;
            element.style.height = Math.min(progress / duration * height, height) + 'px';

            if (progress < duration) {
                window.requestAnimationFrame(animate);
            } else {
                element.style.height = 'auto';
            }
        }
        window.requestAnimationFrame(animate);
    }
};

// 导出动画工具
window.animation = animation; 
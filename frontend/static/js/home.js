// 更新系统状态
function updateSystemStatus() {
    fetch('/api/metrics/latest')
        .then(response => response.json())
        .then(data => {
            // 更新CPU使用率
            updateStatusItem('cpu', data.CPUUsage);
            
            // 更新内存使用率
            updateStatusItem('memory', data.MemoryUsage);
            
            // 更新API健康度
            updateStatusItem('api', data.APIHealth);
        })
        .catch(error => console.error('获取系统指标失败:', error));
}

// 更新状态项
function updateStatusItem(type, value) {
    const item = document.querySelector(`[data-type="${type}"]`);
    if (item) {
        item.querySelector('.progress').style.width = `${value}%`;
        item.querySelector('.status-value').textContent = `${Math.round(value)}%`;
    }
}

// 更新统计卡片
function updateStatCard(cardId, data) {
    const card = document.querySelector(`.stat-card[data-stat="${cardId}"]`);
    if (!card) return;

    // 更新数值
    const numberElement = card.querySelector('.number-animate');
    if (numberElement) {
        animateNumber(numberElement, data.value);
    }

    // 更新趋势
    const trendElement = card.querySelector('.trend');
    if (trendElement) {
        const icon = trendElement.querySelector('i');
        const growth = data.growth;
        
        // 更新图标和颜色
        icon.className = growth > 0 ? 'fas fa-arrow-up' : growth < 0 ? 'fas fa-arrow-down' : 'fas fa-equals';
        trendElement.style.color = growth > 0 ? '#4caf50' : growth < 0 ? '#f44336' : '#9e9e9e';
        
        // 更新百分比文本 - 直接更新 trend 元素的文本
        const text = `${Math.abs(growth).toFixed(1)}%`;
        // 保持图标，更新文本
        trendElement.innerHTML = `<i class="${icon.className}"></i>${text}`;
    }
}

// 数字动画效果
function animateNumber(element, target) {
    const start = parseInt(element.textContent) || 0;
    const duration = 1000;
    const startTime = performance.now();

    function update(currentTime) {
        const elapsed = currentTime - startTime;
        const progress = Math.min(elapsed / duration, 1);

        const current = Math.round(start + (target - start) * progress);
        element.textContent = current.toLocaleString();

        if (progress < 1) {
            requestAnimationFrame(update);
        }
    }

    requestAnimationFrame(update);
}

// 更新统计数据
function updateStats() {
    fetch('/api/stats/dashboard')
        .then(response => response.json())
        .then(data => {
            // 更新今日请求
            updateStatCard('todayRequests', {
                value: data.todayRequests,
                growth: data.requestsGrowth
            });

            // 更新今日对话量
            updateStatCard('todayDialogs', {
                value: data.todayDialogs,
                growth: data.dialogsGrowth
            });

            // 更新平均响应时间
            updateStatCard('avgResponseTime', {
                value: Math.round(data.avgResponseTime),
                growth: -data.responseTimeChange // 响应时间下降为积极趋势
            });

            // 更新系统负载
            updateStatCard('systemLoad', {
                value: Math.round(data.systemLoad), // 直接使用后端返回的值
                growth: 0
            });
        })
        .catch(error => console.error('获取统计数据失败:', error));
}

// 初始化所有图表
let requestMonitorChart = null;
let providerStatsChart = null;

function initCharts() {
    // 初始化Token使用趋势图表
    tokenChart = echarts.init(document.getElementById('tokenTrendChart'));
    modelChart = echarts.init(document.getElementById('modelUsageChart'));
    
    // 更新图表数据
    updateTokenChart('day');
    updateModelChart();

    // 监听时间范围切换
    document.querySelectorAll('.time-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const period = btn.dataset.range;
            // 更新按钮状态
            document.querySelectorAll('.time-btn').forEach(b => {
                b.classList.remove('active');
            });
            btn.classList.add('active');
            // 更新图表
            updateTokenChart(period);
        });
    });

    // 初始化新图表
    requestMonitorChart = echarts.init(document.getElementById('requestMonitorChart'));
    providerStatsChart = echarts.init(document.getElementById('providerStatsChart'));
    
    // 更新数据
    updateRequestMonitor();
    updateProviderStats('day');
    
    // 监听服务商统计时间范围切换
    document.querySelectorAll('.provider-stats .time-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const period = btn.dataset.range;
            updateProviderStats(period);
            
            // 更新按钮状态
            document.querySelectorAll('.provider-stats .time-btn').forEach(b => {
                b.classList.remove('active');
            });
            btn.classList.add('active');
        });
    });
}

// 更新Token使用趋势图表
function updateTokenChart(period) {
    fetch(`/api/stats/tokens?period=${period}`)
        .then(response => response.json())
        .then(data => {
            const chartData = Array.isArray(data) ? data : [];
            
            const option = {
                tooltip: {
                    trigger: 'axis',
                    axisPointer: { type: 'shadow' }
                },
                legend: {
                    data: ['提示词', '回复', '总计'],
                    textStyle: { color: '#fff', fontSize: 10 },
                    right: '4%',
                    top: '0',
                    itemWidth: 12,
                    itemHeight: 8
                },
                grid: {
                    top: '30px',
                    left: '3%',
                    right: '4%',
                    bottom: '12px',
                    containLabel: true
                },
                xAxis: {
                    type: 'category',
                    data: chartData.map(item => item.date),
                    axisLabel: {
                        color: '#fff',
                        fontSize: 9,
                        interval: 'auto'
                    }
                },
                yAxis: {
                    type: 'value',
                    axisLabel: {
                        color: '#fff',
                        fontSize: 9
                    },
                    splitLine: {
                        lineStyle: {
                            color: '#2d2d3f'
                        }
                    }
                },
                series: [
                    {
                        name: '提示词',
                        type: 'bar',
                        stack: 'total',
                        data: chartData.map(item => item.promptTokens),
                        barWidth: '40%'
                    },
                    {
                        name: '回复',
                        type: 'bar',
                        stack: 'total',
                        data: chartData.map(item => item.completionTokens)
                    },
                    {
                        name: '总计',
                        type: 'line',
                        data: chartData.map(item => item.totalTokens),
                        symbolSize: 6
                    }
                ]
            };
            tokenChart.setOption(option);
        });
}

// 更新模型使用分布图表
function updateModelChart() {
    fetch('/api/stats/model-usage')
        .then(response => response.json())
        .then(data => {
            const chartData = Array.isArray(data) ? data : [];
            
            const option = {
                tooltip: {
                    trigger: 'item',
                    formatter: '{a} <br/>{b}: {c} ({d}%)'
                },
                legend: {
                    orient: 'horizontal',
                    top: '0',
                    left: 'center',
                    textStyle: {
                        color: '#fff',
                        fontSize: 9
                    },
                    itemWidth: 12,
                    itemHeight: 8
                },
                series: [
                    {
                        name: '模型使用',
                        type: 'pie',
                        radius: ['40%', '65%'],
                        center: ['50%', '55%'],
                        avoidLabelOverlap: false,
                        itemStyle: {
                            borderRadius: 4,
                            borderColor: '#fff',
                            borderWidth: 1
                        },
                        label: {
                            show: false
                        },
                        emphasis: {
                            label: {
                                show: true,
                                fontSize: 12,
                                fontWeight: 'bold'
                            }
                        },
                        labelLine: {
                            show: false
                        },
                        data: chartData.map(item => ({
                            name: item.modelName,
                            value: item.usage
                        }))
                    }
                ]
            };
            modelChart.setOption(option);
        })
        .catch(error => {
            console.error('获取模型使用数据失败:', error);
            modelChart.setOption({
                series: [{ data: [] }]
            });
        });
}

// 更新实时请求监控
function updateRequestMonitor() {
    fetch('/api/stats/request-monitor')
        .then(response => response.json())
        .then(data => {
            const chartData = Array.isArray(data) ? data : [];
            
            const option = {
                tooltip: {
                    trigger: 'axis',
                    axisPointer: { type: 'cross' }
                },
                legend: {
                    data: ['总请求', '成功', '失败', '响应时间'],
                    textStyle: { color: '#fff', fontSize: 10 },
                    top: '0',
                    right: '4%'
                },
                grid: {
                    top: '30px',
                    left: '3%',
                    right: '4%',
                    bottom: '12px',
                    containLabel: true
                },
                xAxis: {
                    type: 'category',
                    data: chartData.map(item => item.time),
                    axisLabel: {
                        color: '#fff',
                        fontSize: 9
                    }
                },
                yAxis: [
                    {
                        type: 'value',
                        name: '请求数',
                        axisLabel: {
                            color: '#fff',
                            fontSize: 9
                        },
                        splitLine: {
                            lineStyle: { color: '#2d2d3f' }
                        }
                    },
                    {
                        type: 'value',
                        name: '响应时间',
                        axisLabel: {
                            color: '#fff',
                            fontSize: 9,
                            formatter: '{value} ms'
                        },
                        splitLine: { show: false }
                    }
                ],
                series: [
                    {
                        name: '总请求',
                        type: 'line',
                        smooth: true,
                        data: chartData.map(item => item.count || 0)
                    },
                    {
                        name: '成功',
                        type: 'bar',
                        stack: 'total',
                        data: chartData.map(item => item.success || 0)
                    },
                    {
                        name: '失败',
                        type: 'bar',
                        stack: 'total',
                        data: chartData.map(item => item.failed || 0)
                    },
                    {
                        name: '响应时间',
                        type: 'line',
                        yAxisIndex: 1,
                        smooth: true,
                        data: chartData.map(item => item.avgTime || 0)
                    }
                ]
            };
            requestMonitorChart.setOption(option);
        })
        .catch(error => {
            console.error('获取请求监控数据失败:', error);
            requestMonitorChart.setOption({
                series: [{ data: [] }, { data: [] }, { data: [] }, { data: [] }]
            });
        });
}

// 更新服务商统计
function updateProviderStats(period) {
    fetch(`/api/stats/provider-stats?period=${period}`)
        .then(response => response.json())
        .then(data => {
            const chartData = Array.isArray(data) ? data : [];
            
            const option = {
                tooltip: {
                    trigger: 'item'
                },
                legend: {
                    orient: 'horizontal',
                    top: '0',
                    left: 'center',
                    textStyle: {
                        color: '#fff',
                        fontSize: 9
                    }
                },
                series: [
                    {
                        name: '服务商分布',
                        type: 'pie',
                        radius: ['40%', '65%'],
                        center: ['30%', '55%'],
                        data: chartData.map(item => ({
                            name: item.providerName || '未知',
                            value: item.usage || 0
                        }))
                    },
                    {
                        name: '成功率',
                        type: 'gauge',
                        center: ['80%', '55%'],
                        radius: '60%',
                        startAngle: 180,
                        endAngle: 0,
                        min: 0,
                        max: 100,
                        itemStyle: {
                            color: '#36cfc9'
                        },
                        progress: {
                            show: true,
                            roundCap: true,
                            width: 15
                        },
                        pointer: {
                            show: false
                        },
                        axisLine: {
                            roundCap: true,
                            lineStyle: {
                                width: 15
                            }
                        },
                        axisTick: {
                            show: false
                        },
                        splitLine: {
                            show: false
                        },
                        axisLabel: {
                            show: false
                        },
                        title: {
                            show: false
                        },
                        detail: {
                            valueAnimation: true,
                            offsetCenter: [0, '0%'],
                            fontSize: 20,
                            color: '#fff',
                            formatter: '{value}%'
                        },
                        data: [{
                            value: chartData.length > 0 
                                ? chartData.reduce((acc, curr) => acc + (curr.successRate || 0), 0) / chartData.length 
                                : 0
                        }]
                    }
                ]
            };
            providerStatsChart.setOption(option);
        })
        .catch(error => {
            console.error('获取服务商统计数据失败:', error);
            providerStatsChart.setOption({
                series: [{ data: [] }, { data: [{ value: 0 }] }]
            });
        });
}

// 在窗口大小改变时调整图表大小
window.addEventListener('resize', () => {
    tokenChart && tokenChart.resize();
    modelChart && modelChart.resize();
    requestMonitorChart && requestMonitorChart.resize();
    providerStatsChart && providerStatsChart.resize();
});

// 页面加载时启动定时更新
document.addEventListener('DOMContentLoaded', () => {
    updateSystemStatus(); // 更新系统状态
    updateStats(); // 更新统计数据
    initCharts(); // 初始化图表
    updateTokenRanking();
    
    // 定时更新
    setInterval(() => {
        updateSystemStatus();
        updateStats();
        
        // 修复：添加判断，确保元素存在
        const tokenBtn = document.querySelector('.chart-controls .time-btn.active');
        if (tokenBtn) {
            updateTokenChart(tokenBtn.dataset.range);
        }
        
        updateModelChart();
        updateRequestMonitor();
        
        // 修复：添加判断，确保元素存在
        const providerBtn = document.querySelector('.provider-stats .time-btn.active');
        if (providerBtn) {
            updateProviderStats(providerBtn.dataset.range);
        }
        
        updateTokenRanking();
    }, 60000);

    // 添加刷新按钮点击事件
    const refreshBtn = document.querySelector('.token-ranking-card .refresh-btn');
    if (refreshBtn) {
        refreshBtn.addEventListener('click', () => {
            const icon = refreshBtn.querySelector('i');
            icon.classList.add('fa-spin');
            updateTokenRanking().finally(() => {
                setTimeout(() => {
                    icon.classList.remove('fa-spin');
                }, 500);
            });
        });
    }

    // 修改全屏功能
    const fullscreenBtn = document.querySelector('.fullscreen-btn');
    const dashboardContainer = document.querySelector('.dashboard-container');

    if (fullscreenBtn && dashboardContainer) {
        fullscreenBtn.addEventListener('click', () => {
            // 切换全屏类
            dashboardContainer.classList.toggle('fullscreen');
            fullscreenBtn.classList.toggle('active');
            
            // 切换图标
            const icon = fullscreenBtn.querySelector('i');
            icon.classList.toggle('fa-expand');
            icon.classList.toggle('fa-compress');
            
            // 重新调整所有图表大小
            setTimeout(() => {
                tokenChart && tokenChart.resize();
                modelChart && modelChart.resize();
                requestMonitorChart && requestMonitorChart.resize();
                providerStatsChart && providerStatsChart.resize();
            }, 100);
        });
        
        // ESC 键退出全屏
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && dashboardContainer.classList.contains('fullscreen')) {
                dashboardContainer.classList.remove('fullscreen');
                fullscreenBtn.classList.remove('active');
                const icon = fullscreenBtn.querySelector('i');
                icon.classList.add('fa-expand');
                icon.classList.remove('fa-compress');
                
                // 重新调整所有图表大小
                setTimeout(() => {
                    tokenChart && tokenChart.resize();
                    modelChart && modelChart.resize();
                    requestMonitorChart && requestMonitorChart.resize();
                    providerStatsChart && providerStatsChart.resize();
                }, 100);
            }
        });
    }
});

// 更新Token消耗排行
function updateTokenRanking() {
    fetch('/api/stats/token-ranking')
        .then(response => response.json())
        .then(data => {
            const tbody = document.querySelector('.token-ranking tbody');
            if (!tbody) return;

            if (!data || data.length === 0) {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="4" class="empty-state">
                            暂无消耗数据
                        </td>
                    </tr>`;
                return;
            }

            tbody.innerHTML = '';
            data.forEach((item, index) => {
                const tr = document.createElement('tr');
                tr.innerHTML = `
                    <td>
                        <span class="rank-num ${index < 3 ? 'top-' + (index + 1) : ''}">${index + 1}</span>
                        ${item.username}
                    </td>
                    <td>${item.totalTokens.toLocaleString()}</td>
                    <td>${item.percent.toFixed(1)}%</td>
                    <td>
                        <span class="trend ${item.growth > 0 ? 'up' : item.growth < 0 ? 'down' : 'stable'}">
                            <i class="fas fa-${item.growth > 0 ? 'arrow-up' : item.growth < 0 ? 'arrow-down' : 'equals'}"></i>
                            ${Math.abs(item.growth || 0).toFixed(1)}%
                        </span>
                    </td>
                `;
                tbody.appendChild(tr);
            });
        })
        .catch(error => {
            console.error('获取Token排行数据失败:', error);
            const tbody = document.querySelector('.token-ranking tbody');
            if (tbody) {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="4" class="empty-state">
                            数据加载失败
                        </td>
                    </tr>`;
            }
        });
}

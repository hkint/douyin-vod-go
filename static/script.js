document.addEventListener('DOMContentLoaded', () => {
    const urlInput = document.getElementById('douyin-url');
    const parseBtn = document.getElementById('parse-btn');
    const loadingDiv = document.getElementById('loading');
    const resultDiv = document.getElementById('result');
    const resultContent = document.getElementById('result-content');

    const douyinUrlRegex = /((https?:\/\/)?([a-zA-Z0-9-]+\.)*douyin\.com[^\s]*)/;
    
    function formatNumber(num) {
        if (num === null || num === undefined) {
            return '0';
        }
        if (num < 1000) {
            return num.toString();
        } else if (num < 10000) {
            return (num / 1000).toFixed(1).replace(/\.0$/, '') + '千';
        } else if (num < 100000000) {
            return (num / 10000).toFixed(1).replace(/\.0$/, '') + '万';
        } else {
            return (num / 100000000).toFixed(1).replace(/\.0$/, '') + '亿';
        }
    }

    // 将核心解析逻辑封装成一个函数
    async function handleParse() {
        const rawText = urlInput.value.trim();
        if (!rawText) {
            alert('请输入分享内容！');
            return;
        }

        const match = rawText.match(douyinUrlRegex);
        if (!match) {
            displayError("未能从您输入的内容中找到有效的抖音链接，请检查后重试。");
            return;
        }
        const extractedUrl = match[0];
        urlInput.value = extractedUrl;

        loadingDiv.classList.remove('hidden');
        resultDiv.classList.add('hidden');
        resultContent.innerHTML = '';

        try {
            const response = await fetch(`/api/okhk?data&url=${encodeURIComponent(extractedUrl)}`);
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || '解析失败，服务器返回错误');
            }
            const data = await response.json();
            displayResult(data);
        } catch (error) {
            displayError(error.message);
        } finally {
            loadingDiv.classList.add('hidden');
        }
    }

    // 为按钮绑定点击事件
    parseBtn.addEventListener('click', handleParse);

    // --- 新增：为输入框绑定回车键事件 ---
    urlInput.addEventListener('keyup', (event) => {
        // 检查按下的键是否是 "Enter"
        if (event.key === 'Enter') {
            // 阻止默认的回车行为（比如表单提交）
            event.preventDefault();
            // 触发解析按钮的点击事件
            parseBtn.click();
        }
    });

    function displayResult(data) {
        resultDiv.classList.remove('hidden');
        let html = '';

        if (data.nickname) {
            html += `<p><strong>作者：</strong> ${data.nickname}</p>`;
        }
        if (data.desc) {
            html += `<p><strong>标题：</strong> ${data.desc}</p>`;
        }


        if (data.type === 'video' && data.video_url) {
            html += `<p><strong>视频链接：</strong> <a href="${data.video_url}" target="_blank">点击跳转/播放</a></p>`;
            html += `<video controls width="100%" src="${data.video_url}" style="margin-top: 15px; border-radius: 8px;"></video>`;
        } else if (data.type === 'img' && data.image_url_list && data.image_url_list.length > 0) {
            html += `<p><strong>图文集 (共 ${data.image_url_list.length} 张)：</strong></p>`;
            html += '<div class="image-gallery">';
            data.image_url_list.forEach(imgUrl => {
                html += `<a href="${imgUrl}" target="_blank"><img src="${imgUrl}" alt="抖音图片"></a>`;
            });
            html += '</div>';
        } else {
            html += `<p>未找到有效的视频或图片链接。</p>`;
        }

        html += '<div class="stats-container">';
        if (data.digg_count) {
            html += `<div class="stat-item"><strong>${formatNumber(data.digg_count)}</strong><span>点赞</span></div>`;
        }
        if (data.collect_count) {
            html += `<div class="stat-item"><strong>${formatNumber(data.collect_count)}</strong><span>收藏</span></div>`;
        }
        if (data.comment_count) {
            html += `<div class="stat-item"><strong>${formatNumber(data.comment_count)}</strong><span>评论</span></div>`;
        }
        if (data.share_count) {
            html += `<div class="stat-item"><strong>${formatNumber(data.share_count)}</strong><span>分享</span></div>`;
        }
        html += '</div>';

        resultContent.innerHTML = html;
    }

    function displayError(message) {
        resultDiv.classList.remove('hidden');
        resultContent.innerHTML = `<div class="error-message">${message}</div>`;
    }
});
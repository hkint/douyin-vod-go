package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// 新增：用于从文本中提取抖音分享链接的正则表达式
var douyinUrlRegex = regexp.MustCompile(`((https?:\/\/)?([a-zA-Z0-9-]+\.)*douyin\.com[^\s]*)`)

// apiHandler 处理 api/okhk 收到请求
func apiHandler(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头，允许所有来源的跨域请求，方便前端页面调用
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理浏览器预检请求 (OPTIONS)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 从查询参数中获取原始输入内容
	rawInput := r.URL.Query().Get("url")
	if rawInput == "" {
		http.Error(w, "请提供 'url' 参数", http.StatusBadRequest)
		return
	}

	// --- 新增：从原始输入中提取链接 ---
	inputURL := douyinUrlRegex.FindString(rawInput)
	if inputURL == "" {
		http.Error(w, "未能在输入内容中找到有效的抖音链接", http.StatusBadRequest)
		return
	}
	// ------------------------------------

	log.Printf("收到请求: raw_input='%s', extracted_url='%s'", rawInput, inputURL)

	// 检查是否需要返回完整的JSON数据
	if _, ok := r.URL.Query()["data"]; ok {
		videoInfo, err := GetVideoInfo(inputURL)
		if err != nil {
			log.Printf("解析完整信息失败: %v", err)
			http.Error(w, "解析失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(videoInfo)
	} else {
		// 默认只返回视频链接
		videoURL, err := GetVideoURL(inputURL)
		if err != nil {
			log.Printf("解析视频链接失败: %v", err)
			// 如果只获取视频链接失败，很可能是图文类型，尝试返回提示
			http.Error(w, "解析视频链接失败，请尝试使用 'data' 参数获取图文信息。", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(videoURL))
	}
}

// rootHandler 是一个更智能的根路径处理器
func rootHandler(fs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		match := douyinUrlRegex.FindString(path)

		// 如果匹配成功
		if match != "" {
			var douyinURL string

			// 智能判断并重建完整的抖音链接
			if strings.HasPrefix(match, "http") {
				// 如果匹配结果已经包含了协议头，直接使用
				douyinURL = match
			} else {
				// 否则，为其添加 https:// 协议头
				douyinURL = "https://" + match
			}

			log.Printf("通过路径检测到链接: %s", douyinURL)

			// 调用核心逻辑获取无水印视频URL
			finalVideoURL, err := GetVideoURL(douyinURL)
			if err != nil {
				log.Printf("路径链接解析失败: %v", err)
				http.Error(w, "无法从此链接解析出无水印视频，这可能是一个图文链接，请返回主页使用完整解析功能。", http.StatusNotFound)
				return
			}

			// 将用户重定向到无水印视频链接
			log.Printf("重定向到: %s", finalVideoURL)
			http.Redirect(w, r, finalVideoURL, http.StatusFound)
			return
		}

		// 如果路径中没有抖音链接，则正常提供前端文件服务
		fs.ServeHTTP(w, r)
	}
}

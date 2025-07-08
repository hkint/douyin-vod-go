package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	douyinAPITemplate = "https://www.iesdouyin.com/aweme/v1/play/?video_id=%s&ratio=1080p&line=0"
	userAgent         = "Mozilla/5.0 (Linux; Android 15; SAMSUNG SM-S925U) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/16.0 Chrome/138.0.7204.63 Mobile Safari/537.36"
)

// 定义解析视频链接所需的正则表达式
var (
	videoIDRegex      = regexp.MustCompile(`"video":{"play_addr":{"uri":"([a-z0-9]+)"`)
	statsRegex        = regexp.MustCompile(`"statistics"\s*:\s*\{([\s\S]*?)\},`)
	nicknameRegex     = regexp.MustCompile(`"nickname"\s*:\s*"([^"]+)"`)
	signatureRegex    = regexp.MustCompile(`"signature"\s*:\s*"([^"]+)"`)
	createTimeRegex   = regexp.MustCompile(`"create_time":\s*(\d+)`)
	descRegex         = regexp.MustCompile(`"desc":\s*"([^"]+)"`)
	awemeIDRegex      = regexp.MustCompile(`"aweme_id"\s*:\s*"([^"]+)"`)
	commentCountRegex = regexp.MustCompile(`"comment_count"\s*:\s*(\d+)`)
	diggCountRegex    = regexp.MustCompile(`"digg_count"\s*:\s*(\d+)`)
	shareCountRegex   = regexp.MustCompile(`"share_count"\s*:\s*(\d+)`)
	collectCountRegex = regexp.MustCompile(`"collect_count"\s*:\s*(\d+)`)
	imgURLRegex       = regexp.MustCompile(`{"uri":"[^\s"]+","url_list":\["(https:\/\/p\d{1,2}-sign.douyinpic.com\/.*?)"`)
	imgURIRegex       = regexp.MustCompile(`"uri":"([^\s"]+)","url_list":`)
)

// GetVideoInfo 从抖音链接获取完整的视频/图文信息
func GetVideoInfo(url string) (*DouyinVideoInfo, error) {
	body, err := fetchURL(url)
	if err != nil {
		return nil, fmt.Errorf("获取URL内容失败: %w", err)
	}

	info := &DouyinVideoInfo{}

	// 判断类型是视频还是图文
	videoIDMatch := videoIDRegex.FindStringSubmatch(body)
	if len(videoIDMatch) > 1 {
		info.Type = "video"
		info.VideoURL = fmt.Sprintf(douyinAPITemplate, videoIDMatch[1])
	} else {
		info.Type = "img"
		info.ImageURLList = parseImgList(body)
	}

	// 提取统计数据
	statsBlock := statsRegex.FindString(body)
	if statsBlock != "" {
		info.AwemeID = findStringSubmatch(statsBlock, awemeIDRegex)
		info.CommentCount = findIntSubmatch(statsBlock, commentCountRegex)
		info.DiggCount = findIntSubmatch(statsBlock, diggCountRegex)
		info.ShareCount = findIntSubmatch(statsBlock, shareCountRegex)
		info.CollectCount = findIntSubmatch(statsBlock, collectCountRegex)
	}

	// 提取作者信息
	info.Nickname = unescape(findStringSubmatch(body, nicknameRegex))
	info.Signature = unescape(findStringSubmatch(body, signatureRegex))

	// 提取标题
	info.Desc = unescape(findStringSubmatch(body, descRegex))

	// 提取创建时间
	createTimeVal := findIntSubmatch(body, createTimeRegex)
	if createTimeVal > 0 {
		t := time.Unix(createTimeVal, 0)
		info.CreateTime = t.Format("2006-01-02 15:04:05")
	}

	return info, nil
}

// GetVideoURL 仅获取无水印视频链接
func GetVideoURL(url string) (string, error) {
	body, err := fetchURL(url)
	if err != nil {
		return "", fmt.Errorf("获取URL内容失败: %w", err)
	}

	videoIDMatch := videoIDRegex.FindStringSubmatch(body)
	if len(videoIDMatch) < 2 {
		return "", fmt.Errorf("未找到视频ID")
	}
	return fmt.Sprintf(douyinAPITemplate, videoIDMatch[1]), nil
}

// fetchURL 使用指定的User-Agent获取URL内容
func fetchURL(url string) (string, error) {
	// 如果是短链接，需要获取重定向后的真实地址
	finalURL, err := getRedirectURL(url)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", finalURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

// getRedirectURL 处理抖音的短链接，获取重定向后的长链接
func getRedirectURL(url string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 阻止自动重定向
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		location, err := resp.Location()
		if err != nil {
			return "", err
		}
		return location.String(), nil
	}
	return url, nil // 如果没有重定向，返回原URL
}

// parseImgList 解析图文链接
func parseImgList(body string) []string {
	content := strings.ReplaceAll(body, `\u002F`, "/")

	allMatches := imgURLRegex.FindAllStringSubmatch(content, -1)
	uriMatches := imgURIRegex.FindAllStringSubmatch(content, -1)

	uriSet := make(map[string]bool)
	for _, match := range uriMatches {
		if len(match) > 1 {
			uriSet[match[1]] = true
		}
	}

	var result []string
	seen := make(map[string]bool)
	for _, match := range allMatches {
		if len(match) > 1 {
			url := match[1]
			// 确保URL是唯一的，并且属于我们提取到的URI集合
			for uri := range uriSet {
				if strings.Contains(url, uri) && !seen[url] && !strings.Contains(url, "/obj/") {
					result = append(result, url)
					seen[url] = true
					break
				}
			}
		}
	}
	return result
}

// findStringSubmatch 是一个辅助函数，用于从文本中提取匹配的字符串
func findStringSubmatch(text string, re *regexp.Regexp) string {
	match := re.FindStringSubmatch(text)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

// findIntSubmatch 是一个辅助函数，用于从文本中提取匹配的数字
func findIntSubmatch(text string, re *regexp.Regexp) int64 {
	matchStr := findStringSubmatch(text, re)
	if matchStr != "" {
		val, _ := strconv.ParseInt(matchStr, 10, 64)
		return val
	}
	return 0
}

// unescape 处理JSON字符串中的转义字符，例如 \uXXXX
func unescape(s string) string {
	// Go的json库在解码时会自动处理这些转义，但这里我们是直接从HTML中正则提取的，
	// 需要一个简单的方式来解码。最简单的是用json库本身。
	var decodedStr string
	// 我们给它加上引号，让它成为一个合法的JSON字符串
	err := json.Unmarshal([]byte(`"`+s+`"`), &decodedStr)
	if err != nil {
		return s // 如果解码失败，返回原字符串
	}
	return decodedStr
}

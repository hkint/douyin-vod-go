package main

// DouyinVideoInfo 存储解析后的抖音视频/图文信息
type DouyinVideoInfo struct {
	AwemeID      string   `json:"aweme_id"`      // 作品ID
	CommentCount int64    `json:"comment_count"` // 评论数
	DiggCount    int64    `json:"digg_count"`    // 点赞数
	ShareCount   int64    `json:"share_count"`   // 分享数
	CollectCount int64    `json:"collect_count"` // 收藏数
	Nickname     string   `json:"nickname"`      // 作者昵称
	Signature    string   `json:"signature"`     // 作者签名
	Desc         string   `json:"desc"`          // 标题
	CreateTime   string   `json:"create_time"`   // 创建时间
	VideoURL     string   `json:"video_url"`     // 视频链接
	Type         string   `json:"type"`          // 类型 (video/img)
	ImageURLList []string `json:"image_url_list"`// 图片链接列表
}
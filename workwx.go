package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/silenceper/wechat/v2/cache"
	"github.com/xen0n/go-workwx/v2"
)

// WorkwxConfig 企业微信配置
type WorkwxConfig struct {
	CorpID         string `json:"corpId" mapstructure:"corpId"`                 // 企业ID
	AgentID        int64  `json:"agentId" mapstructure:"agentId"`               // 应用ID
	AgentSecret    string `json:"agentSecret" mapstructure:"agentSecret"`       // 应用Secret
	Token          string `json:"token" mapstructure:"token"`                   // 回调Token
	EncodingAESKey string `json:"encodingAESKey" mapstructure:"encodingAESKey"` // 回调加解密Key
	RedisAddr      string `json:"redisAddr" mapstructure:"redisAddr"`           // Redis地址，可选
}

// WorkwxClient 企业微信客户端
type WorkwxClient struct {
	workwxApp           *workwx.WorkwxApp
	config              *WorkwxConfig
	accessTokenProvider workwx.ITokenProvider
	httpClient          *http.Client
}

// NewWorkwxClient 创建企业微信客户端
func NewWorkwxClient(cfg *WorkwxConfig) *WorkwxClient {
	var opts []workwx.CtorOption
	var myCache cache.Cache

	// 如果配置了 Redis，使用 Redis 缓存 access_token
	if cfg.RedisAddr != "" {
		myCache = cache.NewRedis(context.Background(), &cache.RedisOpts{
			Host:     cfg.RedisAddr,
			Username: os.Getenv("REDIS_USERNAME"),
			Password: os.Getenv("REDIS_PASSWORD"),
		})
	} else {
		myCache = cache.NewMemory()
	}

	accessTokenProvider := NewWorkwxAccessTokenProvider(cfg.CorpID, cfg.AgentSecret, myCache)
	opts = append(opts, workwx.WithAccessTokenProvider(accessTokenProvider))

	wx := workwx.New(cfg.CorpID, opts...)
	workwxApp := wx.WithApp(cfg.AgentSecret, cfg.AgentID)

	return &WorkwxClient{
		workwxApp:           workwxApp,
		config:              cfg,
		accessTokenProvider: accessTokenProvider,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetWorkwxApp 获取企业微信App实例，用于直接调用SDK方法
func (c *WorkwxClient) GetWorkwxApp() *workwx.WorkwxApp {
	return c.workwxApp
}

// GetConfig 获取配置
func (c *WorkwxClient) GetConfig() *WorkwxConfig {
	return c.config
}

// CreateHTTPHandler 创建HTTP处理器用于接收企业微信回调
// 需要实现 workwx.RxMessageHandler 接口
func (c *WorkwxClient) CreateHTTPHandler(handler workwx.RxMessageHandler) (*workwx.HTTPHandler, error) {
	return workwx.NewHTTPHandler(c.config.Token, c.config.EncodingAESKey, handler)
}

// SendTextMessage 发送文本消息
// toUser: 成员ID列表，多个用|分隔
// toParty: 部门ID列表，多个用|分隔
// toTag: 标签ID列表，多个用|分隔
// content: 消息内容
func (c *WorkwxClient) SendTextMessage(toUser, toParty, toTag, content string) error {
	recipient := buildRecipient(toUser, toParty, toTag)
	return c.workwxApp.SendTextMessage(recipient, content, false)
}

// SendMarkdownMessage 发送Markdown消息
func (c *WorkwxClient) SendMarkdownMessage(toUser, toParty, toTag, content string) error {
	recipient := buildRecipient(toUser, toParty, toTag)
	return c.workwxApp.SendMarkdownMessage(recipient, content, false)
}

// SendImageMessage 发送图片消息
// mediaID: 素材ID
func (c *WorkwxClient) SendImageMessage(toUser, toParty, toTag, mediaID string) error {
	recipient := buildRecipient(toUser, toParty, toTag)
	return c.workwxApp.SendImageMessage(recipient, mediaID, false)
}

// SendFileMessage 发送文件消息
// mediaID: 素材ID
func (c *WorkwxClient) SendFileMessage(toUser, toParty, toTag, mediaID string) error {
	recipient := buildRecipient(toUser, toParty, toTag)
	return c.workwxApp.SendFileMessage(recipient, mediaID, false)
}

// SendVoiceMessage 发送语音消息
// mediaID: 素材ID
func (c *WorkwxClient) SendVoiceMessage(toUser, toParty, toTag, mediaID string) error {
	recipient := buildRecipient(toUser, toParty, toTag)
	return c.workwxApp.SendVoiceMessage(recipient, mediaID, false)
}

// SendVideoMessage 发送视频消息
// mediaID: 素材ID
// description: 视频描述
// title: 视频标题
func (c *WorkwxClient) SendVideoMessage(toUser, toParty, toTag, mediaID, description, title string) error {
	recipient := buildRecipient(toUser, toParty, toTag)
	return c.workwxApp.SendVideoMessage(recipient, mediaID, description, title, false)
}

// SendTextCardMessage 发送文本卡片消息
// title: 标题
// description: 描述
// url: 跳转链接
// btnTxt: 按钮文字
func (c *WorkwxClient) SendTextCardMessage(toUser, toParty, toTag, title, description, url, btnTxt string) error {
	recipient := buildRecipient(toUser, toParty, toTag)
	return c.workwxApp.SendTextCardMessage(recipient, title, description, url, btnTxt, false)
}

// SendNewsMessage 发送图文消息
// articles: 图文消息列表
func (c *WorkwxClient) SendNewsMessage(toUser, toParty, toTag string, articles []workwx.Article) error {
	recipient := buildRecipient(toUser, toParty, toTag)
	return c.workwxApp.SendNewsMessage(recipient, articles, false)
}

// SendTaskCardMessage 发送任务卡片消息
// title: 标题
// description: 描述
// url: 跳转链接
// taskID: 任务ID
// btn: 按钮列表
func (c *WorkwxClient) SendTaskCardMessage(toUser, toParty, toTag, title, description, url, taskID string, btn []workwx.TaskCardBtn) error {
	recipient := buildRecipient(toUser, toParty, toTag)
	return c.workwxApp.SendTaskCardMessage(recipient, title, description, url, taskID, btn, false)
}

// buildRecipient 构建收件人对象
func buildRecipient(toUser, toParty, toTag string) *workwx.Recipient {
	recipient := &workwx.Recipient{}

	if toUser != "" {
		recipient.UserIDs = strings.Split(toUser, "|")
	}

	if toParty != "" {
		recipient.PartyIDs = strings.Split(toParty, "|")
	}

	if toTag != "" {
		recipient.TagIDs = strings.Split(toTag, "|")
	}

	return recipient
}

// ==================== 客服发送消息 ====================

// KfSendMessageResponse 客服发送消息响应
type KfSendMessageResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	MsgID   string `json:"msgid"`
}

// KfSendTextMessage 客服发送文本消息
// touser: 客户的 external_userid
// openKfID: 客服账号 ID
// msgID: 消息 ID（可选）
// content: 消息内容
func (c *WorkwxClient) KfSendTextMessage(touser, openKfID, msgID, content string) (*KfSendMessageResponse, error) {
	return c.sendKfMessage(touser, openKfID, msgID, map[string]any{
		"msgtype": "text",
		"text": map[string]any{
			"content": content,
		},
	})
}

// KfSendImageMessage 客服发送图片消息
// mediaID: 图片文件 ID
func (c *WorkwxClient) KfSendImageMessage(touser, openKfID, msgID, mediaID string) (*KfSendMessageResponse, error) {
	return c.sendKfMessage(touser, openKfID, msgID, map[string]any{
		"msgtype": "image",
		"image": map[string]any{
			"media_id": mediaID,
		},
	})
}

// KfSendVoiceMessage 客服发送语音消息
// mediaID: 语音文件 ID
func (c *WorkwxClient) KfSendVoiceMessage(touser, openKfID, msgID, mediaID string) (*KfSendMessageResponse, error) {
	return c.sendKfMessage(touser, openKfID, msgID, map[string]any{
		"msgtype": "voice",
		"voice": map[string]any{
			"media_id": mediaID,
		},
	})
}

// KfSendVideoMessage 客服发送视频消息
// mediaID: 视频媒体文件 ID
func (c *WorkwxClient) KfSendVideoMessage(touser, openKfID, msgID, mediaID string) (*KfSendMessageResponse, error) {
	return c.sendKfMessage(touser, openKfID, msgID, map[string]any{
		"msgtype": "video",
		"video": map[string]any{
			"media_id": mediaID,
		},
	})
}

// KfSendFileMessage 客服发送文件消息
// mediaID: 文件 ID
func (c *WorkwxClient) KfSendFileMessage(touser, openKfID, msgID, mediaID string) (*KfSendMessageResponse, error) {
	return c.sendKfMessage(touser, openKfID, msgID, map[string]any{
		"msgtype": "file",
		"file": map[string]any{
			"media_id": mediaID,
		},
	})
}

// KfSendLinkMessage 客服发送图文链接消息
// title: 标题
// desc: 描述
// url: 点击后跳转的链接
// thumbMediaID: 缩略图的 media_id
func (c *WorkwxClient) KfSendLinkMessage(touser, openKfID, msgID, title, desc, url, thumbMediaID string) (*KfSendMessageResponse, error) {
	msgData := map[string]any{
		"msgtype": "link",
		"link": map[string]any{
			"title":          title,
			"url":            url,
			"thumb_media_id": thumbMediaID,
		},
	}
	if desc != "" {
		msgData["link"].(map[string]any)["desc"] = desc
	}
	return c.sendKfMessage(touser, openKfID, msgID, msgData)
}

// KfSendMiniProgramMessage 客服发送小程序消息
// appID: 小程序 appid
// title: 小程序消息标题
// thumbMediaID: 小程序消息封面的 mediaid
// pagePath: 点击消息卡片后进入的小程序页面路径
func (c *WorkwxClient) KfSendMiniProgramMessage(touser, openKfID, msgID, appID, title, thumbMediaID, pagePath string) (*KfSendMessageResponse, error) {
	msgData := map[string]any{
		"msgtype": "miniprogram",
		"miniprogram": map[string]any{
			"appid":          appID,
			"thumb_media_id": thumbMediaID,
			"pagepath":       pagePath,
		},
	}
	if title != "" {
		msgData["miniprogram"].(map[string]any)["title"] = title
	}
	return c.sendKfMessage(touser, openKfID, msgID, msgData)
}

// KfSendLocationMessage 客服发送地理位置消息
// name: 位置名
// address: 地址详情说明
// latitude: 纬度
// longitude: 经度
func (c *WorkwxClient) KfSendLocationMessage(touser, openKfID, msgID, name, address string, latitude, longitude float64) (*KfSendMessageResponse, error) {
	msgData := map[string]any{
		"msgtype": "location",
		"location": map[string]any{
			"latitude":  latitude,
			"longitude": longitude,
		},
	}
	if name != "" {
		msgData["location"].(map[string]any)["name"] = name
	}
	if address != "" {
		msgData["location"].(map[string]any)["address"] = address
	}
	return c.sendKfMessage(touser, openKfID, msgID, msgData)
}

// KfSendCALinkMessage 客服发送获客链接消息
// linkURL: 通过获客助手创建的获客链接
func (c *WorkwxClient) KfSendCALinkMessage(touser, openKfID, msgID, linkURL string) (*KfSendMessageResponse, error) {
	return c.sendKfMessage(touser, openKfID, msgID, map[string]any{
		"msgtype": "ca_link",
		"ca_link": map[string]any{
			"link_url": linkURL,
		},
	})
}

// sendKfMessage 发送客服消息的通用方法
func (c *WorkwxClient) sendKfMessage(touser, openKfID, msgID string, msgData map[string]any) (*KfSendMessageResponse, error) {
	// 构建请求体
	req := map[string]any{
		"touser":    touser,
		"open_kfid": openKfID,
	}
	if msgID != "" {
		req["msgid"] = msgID
	}

	// 合并消息数据
	for k, v := range msgData {
		req[k] = v
	}

	// 调用企业微信 API
	// 使用 go-workwx 的底层执行方法
	// 由于 go-workwx 没有直接暴露 HTTP 客户端，我们需要手动实现
	// 这里先实现一个简单的 HTTP 调用

	return c.kfSendMessage(req)
}

// kfSendMessage 实际调用企业微信客服发送消息接口
func (c *WorkwxClient) kfSendMessage(req map[string]any) (*KfSendMessageResponse, error) {
	// 由于 go-workwx 没有直接暴露底层的 HTTP 客户端
	// 这里需要手动实现 HTTP 调用
	// 可以使用 c.workwxApp 的内部方法，或者使用标准库

	return c.doKfSendMessage(req)
}

// doKfSendMessage 执行客服发送消息的 HTTP 请求
func (c *WorkwxClient) doKfSendMessage(req map[string]any) (*KfSendMessageResponse, error) {
	// 获取 access_token
	token, err := c.accessTokenProvider.GetToken(context.Background())
	if err != nil {
		return nil, fmt.Errorf("获取 access_token 失败: %w", err)
	}

	// 构建请求 URL
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/kf/send_msg?access_token=%s", token)

	// 序列化请求体
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result KfSendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return &result, fmt.Errorf("企业微信 API 返回错误: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

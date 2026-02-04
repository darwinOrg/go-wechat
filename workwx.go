package wechat

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/silenceper/wechat/v2/cache"
	"github.com/xen0n/go-workwx/v2"
)

// WorkwxConfig 企业微信配置
type WorkwxConfig struct {
	CorpID         string // 企业ID
	AgentID        int64  // 应用ID
	AgentSecret    string // 应用Secret
	Token          string // 回调Token
	EncodingAESKey string // 回调加解密Key
	RedisAddr      string // Redis地址，可选
}

// WorkwxClient 企业微信客户端
type WorkwxClient struct {
	workwxApp *workwx.WorkwxApp
	config    *WorkwxConfig
	cache     cache.Cache
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

	opts = append(opts, workwx.WithAccessTokenProvider(NewWorkwxAccessTokenProvider(myCache)))

	wx := workwx.New(cfg.CorpID, opts...)
	workwxApp := wx.WithApp(cfg.AgentSecret, cfg.AgentID)

	return &WorkwxClient{
		workwxApp: workwxApp,
		config:    cfg,
		cache:     myCache,
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

// ParseWorkwxCallbackURL 解析回调URL参数
// 用于手动处理回调URL验证（不使用HTTPHandler时）
func ParseWorkwxCallbackURL(rawURL string) (msgSignature, timestamp, nonce, echostr string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", "", "", fmt.Errorf("解析URL失败: %w", err)
	}

	query := u.Query()
	return query.Get("msg_signature"), query.Get("timestamp"), query.Get("nonce"), query.Get("echostr"), nil
}

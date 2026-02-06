package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/silenceper/wechat/v2/cache"
	"github.com/xen0n/go-workwx/v2"
)

type workwxAccessTokenProvider struct {
	cache      cache.Cache
	keyPrefix  string
	corpID     string
	secret     string
	httpClient *http.Client
}

// getTokenResponse 企业微信获取 access_token 响应
type getTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// NewWorkwxAccessTokenProvider 创建基于 cache.Cache 的 AccessToken 提供者
func NewWorkwxAccessTokenProvider(corpID, secret string, cache cache.Cache) workwx.ITokenProvider {
	return &workwxAccessTokenProvider{
		cache:     cache,
		keyPrefix: "workwx:access_token:",
		corpID:    corpID,
		secret:    secret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetToken 获取 access_token
// 优先从缓存获取，如果缓存中没有，则从企业微信 API 获取
func (p *workwxAccessTokenProvider) GetToken(_ context.Context) (string, error) {
	key := p.keyPrefix + p.secret

	// 先从缓存获取
	rt := p.cache.Get(key)
	if val, ok := rt.(string); ok && val != "" {
		return val, nil
	}

	// 缓存中没有，从企业微信 API 获取
	token, expiresIn, err := p.fetchAccessToken()
	if err != nil {
		return "", fmt.Errorf("获取 access_token 失败: %w", err)
	}

	// 将 token 存入缓存，提前5分钟过期以避免临界问题
	expiration := time.Duration(expiresIn-300) * time.Second
	if expiration < time.Minute {
		expiration = time.Minute // 至少缓存1分钟
	}

	if err := p.cache.Set(key, token, expiration); err != nil {
		// 设置缓存失败不影响返回 token
		return token, nil
	}

	return token, nil
}

// fetchAccessToken 从企业微信 API 获取 access_token
func (p *workwxAccessTokenProvider) fetchAccessToken() (string, int, error) {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", p.corpID, p.secret)

	resp, err := p.httpClient.Get(url)
	if err != nil {
		return "", 0, fmt.Errorf("请求企业微信 API 失败: %w", err)
	}
	defer resp.Body.Close()

	var result getTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return "", 0, fmt.Errorf("企业微信 API 返回错误: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return result.AccessToken, result.ExpiresIn, nil
}

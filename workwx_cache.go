package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/silenceper/wechat/v2/cache"
	"github.com/xen0n/go-workwx/v2"
)

type workwxAccessTokenProvider struct {
	corpID     string
	secret     string
	tokenUrl   string
	cache      cache.Cache
	cacheKey   string
	httpClient *http.Client
	mu         sync.Mutex
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
		corpID:   corpID,
		secret:   secret,
		tokenUrl: fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", corpID, secret),
		cache:    cache,
		cacheKey: "workwx:access_token:" + secret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetToken 获取 access_token
// 优先从缓存获取，如果缓存中没有，则从企业微信 API 获取
// 使用锁避免并发重复请求企业微信 API
func (p *workwxAccessTokenProvider) GetToken(_ context.Context) (string, error) {
	// 先从缓存获取（无锁，快速路径）
	rt := p.cache.Get(p.cacheKey)
	if val, ok := rt.(string); ok && val != "" {
		return val, nil
	}

	// 缓存中没有，加锁获取
	p.mu.Lock()
	defer p.mu.Unlock()

	// 双重检查：可能在等待锁时，其他 goroutine 已经获取并缓存了 token
	rt = p.cache.Get(p.cacheKey)
	if val, ok := rt.(string); ok && val != "" {
		return val, nil
	}

	// 从企业微信 API 获取
	token, expiresIn, err := p.fetchAccessToken()
	if err != nil {
		return "", fmt.Errorf("获取 access_token 失败: %w", err)
	}

	// 将 token 存入缓存，提前5分钟过期以避免临界问题
	expiration := time.Duration(expiresIn-300) * time.Second
	if expiration < time.Minute {
		expiration = time.Minute // 至少缓存1分钟
	}

	if err := p.cache.Set(p.cacheKey, token, expiration); err != nil {
		// 设置缓存失败不影响返回 token
		return token, nil
	}

	return token, nil
}

// fetchAccessToken 从企业微信 API 获取 access_token
func (p *workwxAccessTokenProvider) fetchAccessToken() (string, int, error) {
	resp, err := p.httpClient.Get(p.tokenUrl)
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

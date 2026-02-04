package wechat

import (
	"context"

	"github.com/silenceper/wechat/v2/cache"
	"github.com/xen0n/go-workwx/v2"
)

type workwxAccessTokenProvider struct {
	cache     cache.Cache
	keyPrefix string
}

// NewWorkwxAccessTokenProvider 创建基于 cache.Cache 的 AccessToken 提供者
func NewWorkwxAccessTokenProvider(cache cache.Cache) workwx.ITokenProvider {
	return &workwxAccessTokenProvider{
		cache:     cache,
		keyPrefix: "workwx:access_token:",
	}
}

// GetToken 获取 access_token
func (p *workwxAccessTokenProvider) GetToken(_ context.Context) (string, error) {
	key := p.keyPrefix + "default"
	rt := p.cache.Get(key)
	if val, ok := rt.(string); ok && val != "" {
		return val, nil
	}

	return "", nil
}

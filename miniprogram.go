package wechat

import (
	"context"
	"os"

	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/credential"
	"github.com/silenceper/wechat/v2/miniprogram"
	"github.com/silenceper/wechat/v2/miniprogram/config"
	"github.com/silenceper/wechat/v2/miniprogram/qrcode"
	"github.com/silenceper/wechat/v2/miniprogram/urllink"
)

type MiniProgramConfig struct {
	AppId          string
	AppSecret      string
	ExpireInterval int
	RedisAddr      string
	EnvVersion     string // 小程序版本：正式版为"release"，体验版为"trial"，开发版为"develop"
}

type MiniProgramClient struct {
	miniProgramIns *miniprogram.MiniProgram
	config         *MiniProgramConfig
}

func NewMiniProgramClient(cfg *MiniProgramConfig) *MiniProgramClient {
	miniCfg := &config.Config{
		AppID:     cfg.AppId,
		AppSecret: cfg.AppSecret,
	}
	if cfg.RedisAddr != "" {
		miniCfg.Cache = cache.NewRedis(context.Background(), &cache.RedisOpts{
			Host:     cfg.RedisAddr,
			Username: os.Getenv("REDIS_USERNAME"),
			Password: os.Getenv("REDIS_PASSWORD"),
		})
	} else {
		miniCfg.Cache = cache.NewMemory()
	}

	wx := wechat.NewWechat()
	miniProgramIns := wx.GetMiniProgram(miniCfg)

	stableAccessTokenHandle := credential.NewStableAccessToken(cfg.AppId, cfg.AppSecret, credential.CacheKeyMiniProgramPrefix, miniCfg.Cache)
	miniProgramIns.SetAccessTokenHandle(stableAccessTokenHandle)

	return &MiniProgramClient{miniProgramIns: miniProgramIns, config: cfg}
}

func (c *MiniProgramClient) GenerateUrlLink(path, query string, expireTime int64) (string, error) {
	ulParams := &urllink.ULParams{
		EnvVersion: c.config.EnvVersion,
	}
	if path != "" {
		ulParams.Path = path
	}
	if query != "" {
		ulParams.Query = query
	}
	if expireTime > 0 {
		ulParams.ExpireType = urllink.ExpireTypeTime
		ulParams.ExpireTime = expireTime
	} else {
		ulParams.ExpireType = urllink.ExpireTypeInterval
		ulParams.ExpireInterval = c.config.ExpireInterval
	}

	return c.miniProgramIns.GetURLLink().Generate(ulParams)
}

func (c *MiniProgramClient) GetWXACodeUnlimit(page, scene string, checkPath bool) ([]byte, error) {
	return c.miniProgramIns.GetQRCode().GetWXACodeUnlimit(qrcode.QRCoder{
		Page:       page,
		Path:       page,
		Scene:      scene,
		CheckPath:  &checkPath,
		EnvVersion: c.config.EnvVersion,
	})
}

package wechat_test

import (
	dgctx "github.com/darwinOrg/go-common/context"
	dglogger "github.com/darwinOrg/go-logger"
	"github.com/darwinOrg/go-wechat"
	"os"
	"testing"
)

func TestGenerateUrlLink(t *testing.T) {
	miniProgramClient := &wechat.MiniProgramClient{
		AppID:        os.Getenv("WX_APP_ID"),
		AppSecret:    os.Getenv("WX_APP_SECRET"),
		ForceRefresh: false,
	}

	query := "mobile=3214"
	expireInterval := 30
	params := &wechat.UrlLinkParams{
		Query:          &query,
		ExpireType:     wechat.ExpireTypeInterval,
		ExpireInterval: &expireInterval,
		EnvVersion:     wechat.EnvVersionDevelop,
	}

	ctx := &dgctx.DgContext{TraceId: "123"}
	link, err := miniProgramClient.GenerateUrlLink(ctx, params)
	if err != nil {
		panic(err)
	}
	dglogger.Info(ctx, link)
}

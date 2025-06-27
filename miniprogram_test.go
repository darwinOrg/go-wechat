package wechat_test

import (
	"github.com/darwinOrg/go-wechat"
	"os"
	"testing"
)

func TestGenerateUrlLink(t *testing.T) {
	miniClient := wechat.NewMiniProgramClient(&wechat.MiniProgramConfig{
		AppId:          os.Getenv("WX_APP_ID"),
		AppSecret:      os.Getenv("WX_APP_SECRET"),
		ExpireInterval: 30,
		EnvVersion:     "release",
	})

	path := "sub/interview/pages/direct-interview/sign-up/index"
	query := "job=10394&expiredAt=xxx"
	link, err := miniClient.GenerateUrlLink(path, query, 0)
	if err != nil {
		panic(err)
	}
	t.Logf("link: %s", link)
}

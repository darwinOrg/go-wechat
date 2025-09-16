package wechat_test

import (
	"os"
	"testing"

	"github.com/darwinOrg/go-wechat"
)

func TestGenerateUrlLink(t *testing.T) {
	miniClient := wechat.NewMiniProgramClient(&wechat.MiniProgramConfig{
		AppId:          os.Getenv("WX_APP_ID"),
		AppSecret:      os.Getenv("WX_APP_SECRET"),
		ExpireInterval: 30,
		EnvVersion:     "release",
	})

	path := "path1/path2/path3"
	query := "key1=value1&key2=value2"
	link, err := miniClient.GenerateUrlLink(path, query, 0)
	if err != nil {
		panic(err)
	}
	t.Logf("link: %s", link)
}

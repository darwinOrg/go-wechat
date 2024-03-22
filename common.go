package wechat

import (
	dgctx "github.com/darwinOrg/go-common/context"
	dgerr "github.com/darwinOrg/go-common/enums/error"
	dghttp "github.com/darwinOrg/go-httpclient"
	dglogger "github.com/darwinOrg/go-logger"
)

const (
	// stableAccessTokenURL 获取稳定版access_token的接口
	stableAccessTokenURL = "https://api.weixin.qq.com/cgi-bin/stable_token"
)

// EnvVersion 要打开的小程序版本。正式版为 "release"，体验版为 "trial"，开发版为 "develop"
type EnvVersion string

const (
	// EnvVersionRelease 正式版
	EnvVersionRelease EnvVersion = "release"
	// EnvVersionTrial 体验版
	EnvVersionTrial EnvVersion = "trial"
	// EnvVersionDevelop 开发版
	EnvVersionDevelop EnvVersion = "develop"
)

// ExpireType 失效类型 (指定时间戳/指定间隔)
type ExpireType int

const (
	// ExpireTypeTime 指定时间戳后失效
	ExpireTypeTime ExpireType = 0

	// ExpireTypeInterval 间隔指定天数后失效
	ExpireTypeInterval ExpireType = 1
)

// CommonError 微信返回的通用错误 json
type CommonError struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (e CommonError) Success() bool {
	return e.ErrCode == 0
}

func (e CommonError) BuildDgError() *dgerr.DgError {
	return &dgerr.DgError{
		Code:    int(e.ErrCode),
		Message: e.ErrMsg,
	}
}

type AccessTokenResponse struct {
	CommonError
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func getStableAccessToken(ctx *dgctx.DgContext, appID, appSecret string, forceRefresh bool) (*AccessTokenResponse, error) {
	params := map[string]any{
		"appid":         appID,
		"secret":        appSecret,
		"grant_type":    "client_credential",
		"force_refresh": forceRefresh,
	}

	resp, err := dghttp.DoPostJsonToStruct[AccessTokenResponse](ctx, stableAccessTokenURL, params, nil)
	if err != nil {
		dglogger.Errorf(ctx, "getStableAccessToken error: %v", err)
		return nil, err
	}
	if !resp.Success() {
		return nil, resp.BuildDgError()
	}

	return resp, nil
}

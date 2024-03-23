package wechat

import (
	dgctx "github.com/darwinOrg/go-common/context"
	dghttp "github.com/darwinOrg/go-httpclient"
	dglogger "github.com/darwinOrg/go-logger"
)

const (
	// stableAccessTokenURL 获取稳定版access_token的接口
	urlLinkURL = "https://api.weixin.qq.com/wxa/generate_urllink?access_token="
)

type MiniProgramClient struct {
	AppID        string
	AppSecret    string
	ForceRefresh bool
}

type UrlLinkParams struct {
	Path           *string    `json:"path,omitempty"`
	Query          *string    `json:"query,omitempty"`
	ExpireType     ExpireType `json:"expire_type"`
	ExpireTime     *int64     `json:"expire_time,omitempty"`
	ExpireInterval *int       `json:"expire_interval,omitempty"`
	EnvVersion     EnvVersion `json:"envVersion"`
}

type UrlLinkResponse struct {
	CommonError
	UrlLink string `json:"url_link"`
}

func (c *MiniProgramClient) GenerateUrlLink(ctx *dgctx.DgContext, params *UrlLinkParams) (string, error) {
	dghttp.SetHttpClient(ctx, dghttp.Client11)
	tokenResp, err := getStableAccessToken(ctx, c.AppID, c.AppSecret, c.ForceRefresh)
	if err != nil {
		return "", err
	}

	resp, err := dghttp.DoPostJsonToStruct[UrlLinkResponse](ctx, urlLinkURL+tokenResp.AccessToken, params, nil)
	if err != nil {
		dglogger.Errorf(ctx, "GenerateUrlLink error, params: %+v, err: %v", params, err)
		return "", err
	}
	if !resp.Success() {
		return "", resp.BuildDgError()
	}

	return resp.UrlLink, nil
}

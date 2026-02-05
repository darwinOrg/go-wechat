package wechat_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/darwinOrg/go-wechat"
	"github.com/xen0n/go-workwx/v2"
)

// TestWorkwxClient_SendTextMessage 测试发送文本消息
func TestWorkwxClient_SendTextMessage(t *testing.T) {
	client := wechat.NewWorkwxClient(newWorkwxConfig())

	err := client.SendTextMessage("test_user_id", "", "", "测试消息")
	if err != nil {
		t.Fatalf("SendTextMessage failed: %v", err)
	}
}

// TestWorkwxClient_CreateHTTPHandler 测试创建HTTP处理器
func TestWorkwxClient_CreateHTTPHandler(t *testing.T) {
	client := wechat.NewWorkwxClient(newWorkwxConfig())

	// 创建一个简单的消息处理器
	handler := &testMessageHandler{client: client}

	// 创建HTTP处理器
	httpHandler, err := client.CreateHTTPHandler(handler)
	if err != nil {
		t.Fatalf("CreateHTTPHandler failed: %v", err)
	}

	if httpHandler == nil {
		t.Fatal("CreateHTTPHandler returned nil")
	}

	// 注册HTTP路由
	//http.Handle("/callback", httpHandler)

	// 启动服务器
	// http.ListenAndServe(":8080", nil)
}

// TestVerifyWorkwxCallbackURL 测试验证企业微信回调URL签名
// 使用企业微信文档中的示例数据进行验证
func TestVerifyWorkwxCallbackURL(t *testing.T) {
	// 示例数据来自企业微信官方文档
	//token := "QDG6eK"
	timestamp := "1409659813"
	nonce := "1372623149"
	echostr := "RypEvHKD8QQKFhvQ6QleEB4J58tiPdvo+rtK1I9qca6aM/wvqnLSV5zEPeusUiX5L5X/0lWfrf0QADHHhGd3QczcdCUpj911L3vg3W/sYYvuJTs3TUUkSUXxaccAS0qhxchrRYt66wiSpGLYL42aM6A8dTT+6k4aSknmPj48kzJs8qLjvd4Xgpue06DOdnLxAUHzM6+kDZ+HMZfJYuR+LtwGc2hgf5gsijff0ekUNXZiqATP7PF5mZxZ3Izoun1s4zG4LUMnvw2r+KqCKIw+3IQH03v+BCA9nMELNqbSf6tiWSrXJB3LAVGUcallcrw8V2t9EL4EhzJWrQUax5wLVMNS0+rUPA3k22Ncx4XXZS9o0MBH27Bo6BpNelZpS+/uh9KsNlY6bHCmJU9p8g7m3fVKn28H3KDYA5Pl/T8Z1ptDAVe0lXdQ2YoyyH2uyPIGHBZZIs2pDBS8R07+qN+E7Q=="
	expectedSignature := "477715d11cdb4164915debcba66cb864d751f3e6"

	client := wechat.NewWorkwxClient(&wechat.WorkwxConfig{
		CorpID:         "wx5823bf96d3bd56c7",
		Token:          "QDG6eK",
		EncodingAESKey: "jWmYm7qr5nMoAUwZRjGtBxmz3KA1tkAj3ykkR6q2B2C",
	})

	msg, receiverID, err := client.VerifyURL(expectedSignature, timestamp, nonce, echostr)
	// 验证正确的签名应该通过
	if err != nil {
		t.Fatalf("VerifyWorkwxCallbackURL failed: %v", err)
	}
	log.Println(string(msg))
	log.Println(receiverID)

	// 验证错误的签名应该失败
	invalidSignature := "invalid_signature"
	msg, receiverID, err = client.VerifyURL(invalidSignature, timestamp, nonce, echostr)
	if err != nil {
		t.Fatalf("VerifyWorkwxCallbackURL failed: %v", err)
	}
}

func newWorkwxConfig() *wechat.WorkwxConfig {
	return &wechat.WorkwxConfig{
		CorpID:         "test_corp_id",
		AgentID:        1000001,
		AgentSecret:    "test_secret",
		Token:          "test_token",
		EncodingAESKey: "kWxPEV2QE6N1q9oGNVb5XQCMO1XIQV0MPPkO5q5Fj5o",
		RedisAddr:      "localhost:6379",
	}
}

// testMessageHandler 消息处理器示例
type testMessageHandler struct {
	client *wechat.WorkwxClient
}

// OnIncomingMessage 实现workwx.RxMessageHandler接口
func (h *testMessageHandler) OnIncomingMessage(msg *workwx.RxMessage) error {
	fmt.Printf("收到消息: %+v\n", msg)

	// 根据消息类型处理
	switch msg.MsgType {
	case workwx.MessageTypeText:
		if text, ok := msg.Text(); ok {
			fmt.Printf("文本消息内容: %s\n", text.GetContent())
			// 这里可以根据消息内容决定是否回复
			// 如果需要回复，可以使用 h.client.SendTextMessage 主动发送
		}
	case workwx.MessageTypeImage:
		if image, ok := msg.Image(); ok {
			fmt.Printf("图片消息media_id: %s\n", image.GetMediaID())
		}
		// 处理其他消息类型...
	}

	return nil
}

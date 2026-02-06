package wechat_test

import (
	"fmt"
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
	handler := &testMessageHandler{}

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
		}
	case workwx.MessageTypeImage:
		if image, ok := msg.Image(); ok {
			fmt.Printf("图片消息media_id: %s\n", image.GetMediaID())
		}
		// 处理其他消息类型...
	}

	return nil
}

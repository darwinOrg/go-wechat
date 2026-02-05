package wechat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

const (
	ValidateSignatureError int = -40001
	DecryptAESError        int = -40007
	IllegalBuffer          int = -40008
	DecodeBase64Error      int = -40010
)

// CryptError 企业微信加解密错误
type CryptError struct {
	ErrCode int
	ErrMsg  string
}

func newCryptError(errCode int, errMsg string) *CryptError {
	return &CryptError{ErrCode: errCode, ErrMsg: errMsg}
}

// VerifyURL 验证企业微信回调URL
// 验证签名并解密 echostr，返回解密后的消息
func (c *WorkwxClient) VerifyURL(msgSignature, timestamp, nonce, echostr string) ([]byte, string, *CryptError) {
	signature := c.calSignature(timestamp, nonce, echostr)

	if strings.Compare(signature, msgSignature) != 0 {
		return nil, "", newCryptError(ValidateSignatureError, "signature not equal")
	}

	plaintext, err := c.cbcDecrypter(echostr)
	if err != nil {
		return nil, "", err
	}

	_, _, msg, receiverID, err := c.parsePlainText(plaintext)
	if err != nil {
		return nil, "", err
	}

	return msg, string(receiverID), nil
}

// calSignature 计算签名
func (c *WorkwxClient) calSignature(timestamp, nonce, data string) string {
	sortArr := []string{c.config.Token, timestamp, nonce, data}
	sort.Strings(sortArr)
	var buffer bytes.Buffer
	for _, value := range sortArr {
		buffer.WriteString(value)
	}

	sha := sha1.New()
	sha.Write(buffer.Bytes())
	signature := fmt.Sprintf("%x", sha.Sum(nil))
	return signature
}

// pKCS7Unpadding 去除 PKCS7 填充
func (c *WorkwxClient) pKCS7Unpadding(plaintext []byte, blockSize int) ([]byte, *CryptError) {
	plaintextLen := len(plaintext)
	if plaintext == nil || plaintextLen == 0 {
		return nil, newCryptError(DecryptAESError, "pKCS7Unpadding error nil or zero")
	}
	if plaintextLen%blockSize != 0 {
		return nil, newCryptError(DecryptAESError, "pKCS7Unpadding text not a multiple of the block size")
	}
	paddingLen := int(plaintext[plaintextLen-1])
	return plaintext[:plaintextLen-paddingLen], nil
}

// cbcDecrypter CBC 解密
func (c *WorkwxClient) cbcDecrypter(base64EncryptMsg string) ([]byte, *CryptError) {
	aesKey, err := base64.StdEncoding.DecodeString(c.config.EncodingAESKey + "=")
	if err != nil {
		return nil, newCryptError(DecodeBase64Error, err.Error())
	}

	encryptMsg, err := base64.StdEncoding.DecodeString(base64EncryptMsg)
	if err != nil {
		return nil, newCryptError(DecodeBase64Error, err.Error())
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, newCryptError(DecryptAESError, err.Error())
	}

	if len(encryptMsg) < aes.BlockSize {
		return nil, newCryptError(DecryptAESError, "encrypt_msg size is not valid")
	}

	iv := aesKey[:aes.BlockSize]

	if len(encryptMsg)%aes.BlockSize != 0 {
		return nil, newCryptError(DecryptAESError, "encrypt_msg not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encryptMsg, encryptMsg)

	return encryptMsg, nil
}

// parsePlainText 解析明文
func (c *WorkwxClient) parsePlainText(plaintext []byte) ([]byte, uint32, []byte, []byte, *CryptError) {
	const blockSize = 32
	plaintext, err := c.pKCS7Unpadding(plaintext, blockSize)
	if err != nil {
		return nil, 0, nil, nil, err
	}

	textLen := uint32(len(plaintext))
	if textLen < 20 {
		return nil, 0, nil, nil, newCryptError(IllegalBuffer, "plain is to small 1")
	}
	random := plaintext[:16]
	msgLen := binary.BigEndian.Uint32(plaintext[16:20])
	if textLen < (20 + msgLen) {
		return nil, 0, nil, nil, newCryptError(IllegalBuffer, "plain is to small 2")
	}

	msg := plaintext[20 : 20+msgLen]
	receiverID := plaintext[20+msgLen:]

	return random, msgLen, msg, receiverID, nil
}

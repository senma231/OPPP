package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

// Challenge 认证挑战
type Challenge struct {
	Nonce     string
	Timestamp int64
}

// Response 认证响应
type Response struct {
	Signature string
	NodeID    string
	Timestamp int64
}

// CreateChallenge 创建认证挑战
func CreateChallenge() *Challenge {
	// 生成随机 nonce
	nonce := make([]byte, 16)
	rand.Read(nonce)

	return &Challenge{
		Nonce:     base64.StdEncoding.EncodeToString(nonce),
		Timestamp: time.Now().Unix(),
	}
}

// CreateResponse 创建认证响应
func CreateResponse(challenge *Challenge, nodeID, secret string) *Response {
	// 计算签名
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(fmt.Sprintf("%s:%s:%d", nodeID, challenge.Nonce, challenge.Timestamp)))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return &Response{
		Signature: signature,
		NodeID:    nodeID,
		Timestamp: time.Now().Unix(),
	}
}

// VerifyResponse 验证认证响应
func VerifyResponse(challenge *Challenge, response *Response, secret string) error {
	// 检查时间戳
	if time.Now().Unix()-challenge.Timestamp > 60 {
		return errors.New("挑战已过期")
	}
	if time.Now().Unix()-response.Timestamp > 60 {
		return errors.New("响应已过期")
	}

	// 计算预期签名
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(fmt.Sprintf("%s:%s:%d", response.NodeID, challenge.Nonce, challenge.Timestamp)))
	expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 验证签名
	if response.Signature != expectedSignature {
		return errors.New("签名无效")
	}

	return nil
}

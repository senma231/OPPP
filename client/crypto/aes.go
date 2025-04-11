package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"
)

// AESEncrypt AES 加密
func AESEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	// 创建 cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建 GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 创建随机数
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// AESDecrypt AES 解密
func AESDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// 创建 cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建 GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 检查长度
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("密文太短")
	}

	// 提取 nonce
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EncryptedConn 加密连接
type EncryptedConn struct {
	conn net.Conn
	key  []byte
}

// NewEncryptedConn 创建加密连接
func NewEncryptedConn(conn net.Conn, key []byte) *EncryptedConn {
	return &EncryptedConn{
		conn: conn,
		key:  key,
	}
}

// Read 读取数据
func (c *EncryptedConn) Read(b []byte) (n int, err error) {
	// 读取加密数据长度
	var length uint16
	if err := binary.Read(c.conn, binary.BigEndian, &length); err != nil {
		return 0, err
	}

	// 读取加密数据
	encryptedData := make([]byte, length)
	if _, err := io.ReadFull(c.conn, encryptedData); err != nil {
		return 0, err
	}

	// 解密数据
	decryptedData, err := AESDecrypt(encryptedData, c.key)
	if err != nil {
		return 0, err
	}

	// 复制解密后的数据
	n = copy(b, decryptedData)
	return n, nil
}

// Write 写入数据
func (c *EncryptedConn) Write(b []byte) (n int, err error) {
	// 加密数据
	encryptedData, err := AESEncrypt(b, c.key)
	if err != nil {
		return 0, err
	}

	// 写入加密数据长度
	if err := binary.Write(c.conn, binary.BigEndian, uint16(len(encryptedData))); err != nil {
		return 0, err
	}

	// 写入加密数据
	if _, err := c.conn.Write(encryptedData); err != nil {
		return 0, err
	}

	return len(b), nil
}

// Close 关闭连接
func (c *EncryptedConn) Close() error {
	return c.conn.Close()
}

// LocalAddr 获取本地地址
func (c *EncryptedConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr 获取远程地址
func (c *EncryptedConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline 设置超时
func (c *EncryptedConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// SetReadDeadline 设置读取超时
func (c *EncryptedConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline 设置写入超时
func (c *EncryptedConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

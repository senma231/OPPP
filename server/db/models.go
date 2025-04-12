package db

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	gorm.Model
	Username    string    `gorm:"size:50;not null;uniqueIndex" json:"username"`
	Password    string    `gorm:"size:100;not null" json:"-"`
	Email       string    `gorm:"size:100;uniqueIndex" json:"email"`
	LastLoginAt time.Time `json:"lastLoginAt"`
	IsAdmin     bool      `gorm:"default:false" json:"isAdmin"`
	Devices     []Device  `gorm:"foreignKey:UserID" json:"devices,omitempty"`
}

// Device 设备模型
type Device struct {
	gorm.Model
	UserID     uint      `gorm:"not null" json:"userId"`
	Name       string    `gorm:"size:50;not null" json:"name"`
	NodeID     string    `gorm:"size:50;not null;uniqueIndex" json:"nodeId"`
	Token      string    `gorm:"size:100;not null" json:"-"`
	Status     string    `gorm:"size:20;default:'offline'" json:"status"`
	NATType    string    `gorm:"size:50" json:"natType"`
	ExternalIP string    `gorm:"size:50" json:"externalIP"`
	LocalIP    string    `gorm:"size:50" json:"localIP"`
	Version    string    `gorm:"size:20" json:"version"`
	OS         string    `gorm:"size:20" json:"os"`
	Arch       string    `gorm:"size:20" json:"arch"`
	LastSeenAt time.Time `json:"lastSeenAt"`
	Apps       []App     `gorm:"foreignKey:DeviceID" json:"apps,omitempty"`
}

// App 应用模型
type App struct {
	gorm.Model
	UserID      uint   `gorm:"not null" json:"userId"`
	DeviceID    uint   `gorm:"not null" json:"deviceId"`
	Name        string `gorm:"size:50;not null" json:"name"`
	Protocol    string `gorm:"size:10;not null" json:"protocol"`
	SrcPort     int    `gorm:"not null" json:"srcPort"`
	PeerNode    string `gorm:"size:50;not null" json:"peerNode"`
	DstPort     int    `gorm:"not null" json:"dstPort"`
	DstHost     string `gorm:"size:50;not null" json:"dstHost"`
	Status      string `gorm:"size:20;default:'stopped'" json:"status"`
	Description string `gorm:"size:200" json:"description"`
}

// Forward 转发规则模型
type Forward struct {
	gorm.Model
	UserID      uint   `gorm:"not null" json:"userId"`
	Protocol    string `gorm:"size:10;not null" json:"protocol"`
	SrcPort     int    `gorm:"not null" json:"srcPort"`
	DstHost     string `gorm:"size:50;not null" json:"dstHost"`
	DstPort     int    `gorm:"not null" json:"dstPort"`
	Description string `gorm:"size:200" json:"description"`
	Enabled     bool   `gorm:"default:false" json:"enabled"`
}

// Connection 连接模型
type Connection struct {
	gorm.Model
	SourceDeviceID uint      `gorm:"not null" json:"sourceDeviceId"`
	TargetDeviceID uint      `gorm:"not null" json:"targetDeviceId"`
	Type           string    `gorm:"size:20;not null" json:"type"`
	Status         string    `gorm:"size:20;not null" json:"status"`
	EstablishedAt  time.Time `json:"establishedAt"`
	LastActiveAt   time.Time `json:"lastActiveAt"`
	BytesSent      uint64    `json:"bytesSent"`
	BytesReceived  uint64    `json:"bytesReceived"`
}

// Stats 统计模型
type Stats struct {
	gorm.Model
	UserID         uint   `gorm:"not null" json:"userId"`
	DeviceID       uint   `json:"deviceId"`
	AppID          uint   `json:"appId"`
	ForwardID      uint   `json:"forwardId"`
	BytesSent      uint64 `json:"bytesSent"`
	BytesReceived  uint64 `json:"bytesReceived"`
	Connections    uint64 `json:"connections"`
	ConnectionTime uint64 `json:"connectionTime"`
}

// Session 会话模型
type Session struct {
	gorm.Model
	UserID       uint      `gorm:"not null" json:"userId"`
	Token        string    `gorm:"size:255;not null;uniqueIndex" json:"token"`
	RefreshToken string    `gorm:"size:255;not null;uniqueIndex" json:"refreshToken"`
	UserAgent    string    `gorm:"size:255" json:"userAgent"`
	IP           string    `gorm:"size:50" json:"ip"`
	ExpiresAt    time.Time `json:"expiresAt"`
	LastActiveAt time.Time `json:"lastActiveAt"`
	Revoked      bool      `gorm:"default:false" json:"revoked"`
}

// TOTP 双因素认证模型
type TOTP struct {
	gorm.Model
	UserID      uint      `gorm:"not null;uniqueIndex" json:"userId"`
	Secret      string    `gorm:"size:100;not null" json:"-"`
	Enabled     bool      `gorm:"default:false" json:"enabled"`
	Verified    bool      `gorm:"default:false" json:"verified"`
	LastUsedAt  time.Time `json:"lastUsedAt"`
	BackupCodes []string  `gorm:"type:text" json:"-"`
}

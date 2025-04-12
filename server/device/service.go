package device

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/senma231/p3/common/errors"
	"github.com/senma231/p3/common/logger"
	"github.com/senma231/p3/server/db"
	"gorm.io/gorm"
)

// Service 设备服务
type Service struct {
}

// NewService 创建设备服务
func NewService() *Service {
	return &Service{}
}

// DeviceRequest 设备请求
type DeviceRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description"`
}

// DeviceUpdateRequest 设备更新请求
type DeviceUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// DeviceStatusRequest 设备状态更新请求
type DeviceStatusRequest struct {
	Status     string `json:"status" binding:"required"`
	NATType    string `json:"natType"`
	ExternalIP string `json:"externalIP"`
	LocalIP    string `json:"localIP"`
	Version    string `json:"version"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
}

// GetDevices 获取用户的所有设备
func (s *Service) GetDevices(userID uint) ([]db.Device, error) {
	var devices []db.Device
	if result := db.DB.Where("user_id = ?", userID).Find(&devices); result.Error != nil {
		return nil, errors.Database("查询设备失败", result.Error)
	}
	return devices, nil
}

// GetDevice 获取设备详情
func (s *Service) GetDevice(userID uint, deviceID uint) (*db.Device, error) {
	var device db.Device
	if result := db.DB.Where("id = ? AND user_id = ?", deviceID, userID).First(&device); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("设备不存在")
		}
		return nil, errors.Database("查询设备失败", result.Error)
	}
	return &device, nil
}

// GetDeviceByNodeID 根据节点 ID 获取设备
func (s *Service) GetDeviceByNodeID(nodeID string) (*db.Device, error) {
	var device db.Device
	if result := db.DB.Where("node_id = ?", nodeID).First(&device); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("设备不存在")
		}
		return nil, errors.Database("查询设备失败", result.Error)
	}
	return &device, nil
}

// CreateDevice 创建设备
func (s *Service) CreateDevice(userID uint, req *DeviceRequest) (*db.Device, error) {
	// 生成节点 ID 和令牌
	nodeID, err := generateNodeID()
	if err != nil {
		return nil, errors.Internal("生成节点 ID 失败")
	}

	token, err := generateToken()
	if err != nil {
		return nil, errors.Internal("生成令牌失败")
	}

	// 创建设备
	device := &db.Device{
		UserID:     userID,
		Name:       req.Name,
		NodeID:     nodeID,
		Token:      token,
		Status:     "offline",
		LastSeenAt: time.Now(),
	}

	if result := db.DB.Create(device); result.Error != nil {
		return nil, errors.Database("创建设备失败", result.Error)
	}

	return device, nil
}

// UpdateDevice 更新设备
func (s *Service) UpdateDevice(userID uint, deviceID uint, req *DeviceUpdateRequest) (*db.Device, error) {
	var device db.Device
	if result := db.DB.Where("id = ? AND user_id = ?", deviceID, userID).First(&device); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("设备不存在")
		}
		return nil, errors.Database("查询设备失败", result.Error)
	}

	// 更新设备信息
	if req.Name != "" {
		device.Name = req.Name
	}

	if result := db.DB.Save(&device); result.Error != nil {
		return nil, errors.Database("更新设备失败", result.Error)
	}

	return &device, nil
}

// DeleteDevice 删除设备
func (s *Service) DeleteDevice(userID uint, deviceID uint) error {
	var device db.Device
	if result := db.DB.Where("id = ? AND user_id = ?", deviceID, userID).First(&device); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.NotFound("设备不存在")
		}
		return errors.Database("查询设备失败", result.Error)
	}

	// 删除设备
	if result := db.DB.Delete(&device); result.Error != nil {
		return errors.Database("删除设备失败", result.Error)
	}

	return nil
}

// UpdateDeviceStatus 更新设备状态
func (s *Service) UpdateDeviceStatus(deviceID uint, req *DeviceStatusRequest) (*db.Device, error) {
	var device db.Device
	if result := db.DB.First(&device, deviceID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("设备不存在")
		}
		return nil, errors.Database("查询设备失败", result.Error)
	}

	// 更新设备状态
	device.Status = req.Status
	device.NATType = req.NATType
	device.ExternalIP = req.ExternalIP
	device.LocalIP = req.LocalIP
	device.Version = req.Version
	device.OS = req.OS
	device.Arch = req.Arch
	device.LastSeenAt = time.Now()

	if result := db.DB.Save(&device); result.Error != nil {
		return nil, errors.Database("更新设备状态失败", result.Error)
	}

	return &device, nil
}

// AuthenticateDevice 设备认证
func (s *Service) AuthenticateDevice(nodeID, token string) (*db.Device, error) {
	var device db.Device
	if result := db.DB.Where("node_id = ?", nodeID).First(&device); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("设备不存在")
		}
		return nil, errors.Database("查询设备失败", result.Error)
	}

	// 验证令牌
	if device.Token != token {
		return nil, errors.Unauthorized("设备令牌无效")
	}

	// 更新设备状态
	device.Status = "online"
	device.LastSeenAt = time.Now()

	if result := db.DB.Save(&device); result.Error != nil {
		logger.Warn("更新设备状态失败: %v", result.Error)
	}

	return &device, nil
}

// RegenerateToken 重新生成设备令牌
func (s *Service) RegenerateToken(userID uint, deviceID uint) (string, error) {
	var device db.Device
	if result := db.DB.Where("id = ? AND user_id = ?", deviceID, userID).First(&device); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", errors.NotFound("设备不存在")
		}
		return "", errors.Database("查询设备失败", result.Error)
	}

	// 生成新令牌
	token, err := generateToken()
	if err != nil {
		return "", errors.Internal("生成令牌失败")
	}

	// 更新设备令牌
	device.Token = token
	if result := db.DB.Save(&device); result.Error != nil {
		return "", errors.Database("更新设备令牌失败", result.Error)
	}

	return token, nil
}

// generateNodeID 生成节点 ID
func generateNodeID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateToken 生成令牌
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

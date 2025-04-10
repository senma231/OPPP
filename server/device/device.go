package device

import (
	"errors"
	"fmt"
	"time"

	"github.com/senma231/p3/server/config"
	"github.com/senma231/p3/server/db"
	"gorm.io/gorm"
)

// Service 设备服务
type Service struct {
	config *config.Config
}

// NewService 创建设备服务
func NewService(cfg *config.Config) *Service {
	return &Service{
		config: cfg,
	}
}

// CreateDevice 创建设备
func (s *Service) CreateDevice(userID uint, name, nodeID, token string) (*db.Device, error) {
	// 检查节点 ID 是否已存在
	var existingDevice db.Device
	if err := db.DB.Where("node_id = ?", nodeID).First(&existingDevice).Error; err == nil {
		return nil, errors.New("节点 ID 已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}

	// 创建设备
	device := &db.Device{
		UserID: userID,
		Name:   name,
		NodeID: nodeID,
		Token:  token,
		Status: "offline",
	}

	if err := db.DB.Create(device).Error; err != nil {
		return nil, fmt.Errorf("创建设备失败: %w", err)
	}

	return device, nil
}

// GetDeviceByID 根据 ID 获取设备
func (s *Service) GetDeviceByID(deviceID uint) (*db.Device, error) {
	var device db.Device
	if err := db.DB.First(&device, deviceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("设备不存在")
		}
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}
	return &device, nil
}

// GetDeviceByNodeID 根据节点 ID 获取设备
func (s *Service) GetDeviceByNodeID(nodeID string) (*db.Device, error) {
	var device db.Device
	if err := db.DB.Where("node_id = ?", nodeID).First(&device).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("设备不存在")
		}
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}
	return &device, nil
}

// GetDevicesByUserID 获取用户的所有设备
func (s *Service) GetDevicesByUserID(userID uint) ([]db.Device, error) {
	var devices []db.Device
	if err := db.DB.Where("user_id = ?", userID).Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}
	return devices, nil
}

// UpdateDevice 更新设备信息
func (s *Service) UpdateDevice(deviceID uint, updates map[string]interface{}) (*db.Device, error) {
	device, err := s.GetDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}

	if err := db.DB.Model(device).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新设备失败: %w", err)
	}

	return device, nil
}

// DeleteDevice 删除设备
func (s *Service) DeleteDevice(deviceID uint) error {
	if err := db.DB.Delete(&db.Device{}, deviceID).Error; err != nil {
		return fmt.Errorf("删除设备失败: %w", err)
	}
	return nil
}

// UpdateDeviceStatus 更新设备状态
func (s *Service) UpdateDeviceStatus(nodeID, status, natType, externalIP, localIP, version, os, arch string) (*db.Device, error) {
	device, err := s.GetDeviceByNodeID(nodeID)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"status":      status,
		"nat_type":    natType,
		"external_ip": externalIP,
		"local_ip":    localIP,
		"version":     version,
		"os":          os,
		"arch":        arch,
		"last_seen_at": time.Now(),
	}

	if err := db.DB.Model(device).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新设备状态失败: %w", err)
	}

	return device, nil
}

// VerifyDeviceToken 验证设备令牌
func (s *Service) VerifyDeviceToken(nodeID, token string) (bool, error) {
	device, err := s.GetDeviceByNodeID(nodeID)
	if err != nil {
		return false, err
	}

	return device.Token == token, nil
}

// GetOnlineDevices 获取在线设备
func (s *Service) GetOnlineDevices() ([]db.Device, error) {
	var devices []db.Device
	if err := db.DB.Where("status = ?", "online").Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}
	return devices, nil
}

// GetDeviceStats 获取设备统计信息
func (s *Service) GetDeviceStats(deviceID uint) (map[string]interface{}, error) {
	// 获取设备
	device, err := s.GetDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}

	// 获取设备的应用数量
	var appCount int64
	if err := db.DB.Model(&db.App{}).Where("device_id = ?", deviceID).Count(&appCount).Error; err != nil {
		return nil, fmt.Errorf("查询应用数量失败: %w", err)
	}

	// 获取设备的连接数量
	var connectionCount int64
	if err := db.DB.Model(&db.Connection{}).Where("source_device_id = ? OR target_device_id = ?", deviceID, deviceID).Count(&connectionCount).Error; err != nil {
		return nil, fmt.Errorf("查询连接数量失败: %w", err)
	}

	// 获取设备的流量统计
	var stats db.Stats
	if err := db.DB.Where("device_id = ?", deviceID).Order("created_at DESC").First(&stats).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("查询统计信息失败: %w", err)
		}
		// 如果没有统计信息，使用默认值
		stats = db.Stats{
			DeviceID:       deviceID,
			BytesSent:      0,
			BytesReceived:  0,
			Connections:    0,
			ConnectionTime: 0,
		}
	}

	// 返回统计信息
	return map[string]interface{}{
		"device":         device,
		"appCount":       appCount,
		"connectionCount": connectionCount,
		"bytesSent":      stats.BytesSent,
		"bytesReceived":  stats.BytesReceived,
		"connections":    stats.Connections,
		"connectionTime": stats.ConnectionTime,
	}, nil
}

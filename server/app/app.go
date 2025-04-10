package app

import (
	"errors"
	"fmt"

	"github.com/senma231/p3/server/config"
	"github.com/senma231/p3/server/db"
	"gorm.io/gorm"
)

// Service 应用服务
type Service struct {
	config *config.Config
}

// NewService 创建应用服务
func NewService(cfg *config.Config) *Service {
	return &Service{
		config: cfg,
	}
}

// CreateApp 创建应用
func (s *Service) CreateApp(userID, deviceID uint, name, protocol string, srcPort int, peerNode string, dstPort int, dstHost, description string) (*db.App, error) {
	// 检查设备是否存在
	var device db.Device
	if err := db.DB.First(&device, deviceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("设备不存在")
		}
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}

	// 检查设备是否属于用户
	if device.UserID != userID {
		return nil, errors.New("设备不属于该用户")
	}

	// 检查源端口是否已被使用
	var existingApp db.App
	if err := db.DB.Where("device_id = ? AND src_port = ?", deviceID, srcPort).First(&existingApp).Error; err == nil {
		return nil, errors.New("源端口已被使用")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询应用失败: %w", err)
	}

	// 创建应用
	app := &db.App{
		UserID:      userID,
		DeviceID:    deviceID,
		Name:        name,
		Protocol:    protocol,
		SrcPort:     srcPort,
		PeerNode:    peerNode,
		DstPort:     dstPort,
		DstHost:     dstHost,
		Status:      "stopped",
		Description: description,
	}

	if err := db.DB.Create(app).Error; err != nil {
		return nil, fmt.Errorf("创建应用失败: %w", err)
	}

	return app, nil
}

// GetAppByID 根据 ID 获取应用
func (s *Service) GetAppByID(appID uint) (*db.App, error) {
	var app db.App
	if err := db.DB.First(&app, appID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("应用不存在")
		}
		return nil, fmt.Errorf("查询应用失败: %w", err)
	}
	return &app, nil
}

// GetAppsByUserID 获取用户的所有应用
func (s *Service) GetAppsByUserID(userID uint) ([]db.App, error) {
	var apps []db.App
	if err := db.DB.Where("user_id = ?", userID).Find(&apps).Error; err != nil {
		return nil, fmt.Errorf("查询应用失败: %w", err)
	}
	return apps, nil
}

// GetAppsByDeviceID 获取设备的所有应用
func (s *Service) GetAppsByDeviceID(deviceID uint) ([]db.App, error) {
	var apps []db.App
	if err := db.DB.Where("device_id = ?", deviceID).Find(&apps).Error; err != nil {
		return nil, fmt.Errorf("查询应用失败: %w", err)
	}
	return apps, nil
}

// UpdateApp 更新应用信息
func (s *Service) UpdateApp(appID uint, updates map[string]interface{}) (*db.App, error) {
	app, err := s.GetAppByID(appID)
	if err != nil {
		return nil, err
	}

	if err := db.DB.Model(app).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新应用失败: %w", err)
	}

	return app, nil
}

// DeleteApp 删除应用
func (s *Service) DeleteApp(appID uint) error {
	if err := db.DB.Delete(&db.App{}, appID).Error; err != nil {
		return fmt.Errorf("删除应用失败: %w", err)
	}
	return nil
}

// StartApp 启动应用
func (s *Service) StartApp(appID uint) (*db.App, error) {
	app, err := s.GetAppByID(appID)
	if err != nil {
		return nil, err
	}

	if app.Status == "running" {
		return app, nil
	}

	if err := db.DB.Model(app).Update("status", "running").Error; err != nil {
		return nil, fmt.Errorf("更新应用状态失败: %w", err)
	}

	return app, nil
}

// StopApp 停止应用
func (s *Service) StopApp(appID uint) (*db.App, error) {
	app, err := s.GetAppByID(appID)
	if err != nil {
		return nil, err
	}

	if app.Status == "stopped" {
		return app, nil
	}

	if err := db.DB.Model(app).Update("status", "stopped").Error; err != nil {
		return nil, fmt.Errorf("更新应用状态失败: %w", err)
	}

	return app, nil
}

// GetAppStats 获取应用统计信息
func (s *Service) GetAppStats(appID uint) (map[string]interface{}, error) {
	// 获取应用
	app, err := s.GetAppByID(appID)
	if err != nil {
		return nil, err
	}

	// 获取应用的统计信息
	var stats db.Stats
	if err := db.DB.Where("app_id = ?", appID).Order("created_at DESC").First(&stats).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("查询统计信息失败: %w", err)
		}
		// 如果没有统计信息，使用默认值
		stats = db.Stats{
			AppID:          appID,
			BytesSent:      0,
			BytesReceived:  0,
			Connections:    0,
			ConnectionTime: 0,
		}
	}

	// 返回统计信息
	return map[string]interface{}{
		"app":            app,
		"bytesSent":      stats.BytesSent,
		"bytesReceived":  stats.BytesReceived,
		"connections":    stats.Connections,
		"connectionTime": stats.ConnectionTime,
	}, nil
}

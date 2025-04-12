package app

import (
	"github.com/senma231/p3/common/errors"
	"github.com/senma231/p3/server/db"
	"gorm.io/gorm"
)

// Service 应用服务
type Service struct {
}

// NewService 创建应用服务
func NewService() *Service {
	return &Service{}
}

// AppRequest 应用请求
type AppRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Protocol    string `json:"protocol" binding:"required,oneof=tcp udp"`
	SrcPort     int    `json:"srcPort" binding:"required,min=1,max=65535"`
	PeerNode    string `json:"peerNode" binding:"required"`
	DstPort     int    `json:"dstPort" binding:"required,min=1,max=65535"`
	DstHost     string `json:"dstHost" binding:"required"`
	Description string `json:"description"`
}

// AppUpdateRequest 应用更新请求
type AppUpdateRequest struct {
	Name        string `json:"name"`
	Protocol    string `json:"protocol" binding:"omitempty,oneof=tcp udp"`
	SrcPort     int    `json:"srcPort" binding:"omitempty,min=1,max=65535"`
	PeerNode    string `json:"peerNode"`
	DstPort     int    `json:"dstPort" binding:"omitempty,min=1,max=65535"`
	DstHost     string `json:"dstHost"`
	Description string `json:"description"`
}

// GetApps 获取用户的所有应用
func (s *Service) GetApps(userID uint) ([]db.App, error) {
	var apps []db.App
	if result := db.DB.Where("user_id = ?", userID).Find(&apps); result.Error != nil {
		return nil, errors.Database("查询应用失败", result.Error)
	}
	return apps, nil
}

// GetApp 获取应用详情
func (s *Service) GetApp(userID uint, appID uint) (*db.App, error) {
	var app db.App
	if result := db.DB.Where("id = ? AND user_id = ?", appID, userID).First(&app); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("应用不存在")
		}
		return nil, errors.Database("查询应用失败", result.Error)
	}
	return &app, nil
}

// CreateApp 创建应用
func (s *Service) CreateApp(userID uint, deviceID uint, req *AppRequest) (*db.App, error) {
	// 检查设备是否存在
	var device db.Device
	if result := db.DB.Where("id = ? AND user_id = ?", deviceID, userID).First(&device); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("设备不存在")
		}
		return nil, errors.Database("查询设备失败", result.Error)
	}

	// 检查对等节点是否存在
	var peerDevice db.Device
	if result := db.DB.Where("node_id = ?", req.PeerNode).First(&peerDevice); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("对等节点不存在")
		}
		return nil, errors.Database("查询对等节点失败", result.Error)
	}

	// 检查端口是否已被使用
	var existingApp db.App
	if result := db.DB.Where("device_id = ? AND src_port = ?", deviceID, req.SrcPort).First(&existingApp); result.Error == nil {
		return nil, errors.Conflict("端口已被使用")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.Database("查询应用失败", result.Error)
	}

	// 创建应用
	app := &db.App{
		UserID:      userID,
		DeviceID:    deviceID,
		Name:        req.Name,
		Protocol:    req.Protocol,
		SrcPort:     req.SrcPort,
		PeerNode:    req.PeerNode,
		DstPort:     req.DstPort,
		DstHost:     req.DstHost,
		Status:      "stopped",
		Description: req.Description,
	}

	if result := db.DB.Create(app); result.Error != nil {
		return nil, errors.Database("创建应用失败", result.Error)
	}

	return app, nil
}

// UpdateApp 更新应用
func (s *Service) UpdateApp(userID uint, appID uint, req *AppUpdateRequest) (*db.App, error) {
	var app db.App
	if result := db.DB.Where("id = ? AND user_id = ?", appID, userID).First(&app); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("应用不存在")
		}
		return nil, errors.Database("查询应用失败", result.Error)
	}

	// 更新应用信息
	if req.Name != "" {
		app.Name = req.Name
	}
	if req.Protocol != "" {
		app.Protocol = req.Protocol
	}
	if req.SrcPort > 0 {
		// 检查端口是否已被使用
		var existingApp db.App
		if result := db.DB.Where("device_id = ? AND src_port = ? AND id != ?", app.DeviceID, req.SrcPort, appID).First(&existingApp); result.Error == nil {
			return nil, errors.Conflict("端口已被使用")
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Database("查询应用失败", result.Error)
		}
		app.SrcPort = req.SrcPort
	}
	if req.PeerNode != "" {
		// 检查对等节点是否存在
		var peerDevice db.Device
		if result := db.DB.Where("node_id = ?", req.PeerNode).First(&peerDevice); result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil, errors.NotFound("对等节点不存在")
			}
			return nil, errors.Database("查询对等节点失败", result.Error)
		}
		app.PeerNode = req.PeerNode
	}
	if req.DstPort > 0 {
		app.DstPort = req.DstPort
	}
	if req.DstHost != "" {
		app.DstHost = req.DstHost
	}
	if req.Description != "" {
		app.Description = req.Description
	}

	if result := db.DB.Save(&app); result.Error != nil {
		return nil, errors.Database("更新应用失败", result.Error)
	}

	return &app, nil
}

// DeleteApp 删除应用
func (s *Service) DeleteApp(userID uint, appID uint) error {
	var app db.App
	if result := db.DB.Where("id = ? AND user_id = ?", appID, userID).First(&app); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.NotFound("应用不存在")
		}
		return errors.Database("查询应用失败", result.Error)
	}

	// 删除应用
	if result := db.DB.Delete(&app); result.Error != nil {
		return errors.Database("删除应用失败", result.Error)
	}

	return nil
}

// StartApp 启动应用
func (s *Service) StartApp(userID uint, appID uint) (*db.App, error) {
	var app db.App
	if result := db.DB.Where("id = ? AND user_id = ?", appID, userID).First(&app); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("应用不存在")
		}
		return nil, errors.Database("查询应用失败", result.Error)
	}

	// 检查应用状态
	if app.Status == "running" {
		return nil, errors.Conflict("应用已在运行")
	}

	// 更新应用状态
	app.Status = "running"
	if result := db.DB.Save(&app); result.Error != nil {
		return nil, errors.Database("更新应用状态失败", result.Error)
	}

	return &app, nil
}

// StopApp 停止应用
func (s *Service) StopApp(userID uint, appID uint) (*db.App, error) {
	var app db.App
	if result := db.DB.Where("id = ? AND user_id = ?", appID, userID).First(&app); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("应用不存在")
		}
		return nil, errors.Database("查询应用失败", result.Error)
	}

	// 检查应用状态
	if app.Status == "stopped" {
		return nil, errors.Conflict("应用已停止")
	}

	// 更新应用状态
	app.Status = "stopped"
	if result := db.DB.Save(&app); result.Error != nil {
		return nil, errors.Database("更新应用状态失败", result.Error)
	}

	return &app, nil
}

// GetAppsByDevice 获取设备的所有应用
func (s *Service) GetAppsByDevice(deviceID uint) ([]db.App, error) {
	var apps []db.App
	if result := db.DB.Where("device_id = ?", deviceID).Find(&apps); result.Error != nil {
		return nil, errors.Database("查询应用失败", result.Error)
	}
	return apps, nil
}

// GetAppsByPeerNode 获取对等节点的所有应用
func (s *Service) GetAppsByPeerNode(peerNode string) ([]db.App, error) {
	var apps []db.App
	if result := db.DB.Where("peer_node = ?", peerNode).Find(&apps); result.Error != nil {
		return nil, errors.Database("查询应用失败", result.Error)
	}
	return apps, nil
}

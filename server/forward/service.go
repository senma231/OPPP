package forward

import (
	"github.com/senma231/p3/common/errors"
	"github.com/senma231/p3/server/db"
	"gorm.io/gorm"
)

// Service 转发服务
type Service struct {
}

// NewService 创建转发服务
func NewService() *Service {
	return &Service{}
}

// ForwardRequest 转发请求
type ForwardRequest struct {
	Protocol    string `json:"protocol" binding:"required,oneof=tcp udp"`
	SrcPort     int    `json:"srcPort" binding:"required,min=1,max=65535"`
	DstHost     string `json:"dstHost" binding:"required"`
	DstPort     int    `json:"dstPort" binding:"required,min=1,max=65535"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// ForwardUpdateRequest 转发更新请求
type ForwardUpdateRequest struct {
	Protocol    string `json:"protocol" binding:"omitempty,oneof=tcp udp"`
	SrcPort     int    `json:"srcPort" binding:"omitempty,min=1,max=65535"`
	DstHost     string `json:"dstHost"`
	DstPort     int    `json:"dstPort" binding:"omitempty,min=1,max=65535"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`
}

// GetForwards 获取用户的所有转发规则
func (s *Service) GetForwards(userID uint) ([]db.Forward, error) {
	var forwards []db.Forward
	if result := db.DB.Where("user_id = ?", userID).Find(&forwards); result.Error != nil {
		return nil, errors.Database("查询转发规则失败", result.Error)
	}
	return forwards, nil
}

// GetForward 获取转发规则详情
func (s *Service) GetForward(userID uint, forwardID uint) (*db.Forward, error) {
	var forward db.Forward
	if result := db.DB.Where("id = ? AND user_id = ?", forwardID, userID).First(&forward); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("转发规则不存在")
		}
		return nil, errors.Database("查询转发规则失败", result.Error)
	}
	return &forward, nil
}

// CreateForward 创建转发规则
func (s *Service) CreateForward(userID uint, req *ForwardRequest) (*db.Forward, error) {
	// 检查端口是否已被使用
	var existingForward db.Forward
	if result := db.DB.Where("user_id = ? AND src_port = ?", userID, req.SrcPort).First(&existingForward); result.Error == nil {
		return nil, errors.Conflict("端口已被使用")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.Database("查询转发规则失败", result.Error)
	}

	// 创建转发规则
	forward := &db.Forward{
		UserID:      userID,
		Protocol:    req.Protocol,
		SrcPort:     req.SrcPort,
		DstHost:     req.DstHost,
		DstPort:     req.DstPort,
		Description: req.Description,
		Enabled:     req.Enabled,
	}

	if result := db.DB.Create(forward); result.Error != nil {
		return nil, errors.Database("创建转发规则失败", result.Error)
	}

	return forward, nil
}

// UpdateForward 更新转发规则
func (s *Service) UpdateForward(userID uint, forwardID uint, req *ForwardUpdateRequest) (*db.Forward, error) {
	var forward db.Forward
	if result := db.DB.Where("id = ? AND user_id = ?", forwardID, userID).First(&forward); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("转发规则不存在")
		}
		return nil, errors.Database("查询转发规则失败", result.Error)
	}

	// 更新转发规则信息
	if req.Protocol != "" {
		forward.Protocol = req.Protocol
	}
	if req.SrcPort > 0 {
		// 检查端口是否已被使用
		var existingForward db.Forward
		if result := db.DB.Where("user_id = ? AND src_port = ? AND id != ?", userID, req.SrcPort, forwardID).First(&existingForward); result.Error == nil {
			return nil, errors.Conflict("端口已被使用")
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Database("查询转发规则失败", result.Error)
		}
		forward.SrcPort = req.SrcPort
	}
	if req.DstHost != "" {
		forward.DstHost = req.DstHost
	}
	if req.DstPort > 0 {
		forward.DstPort = req.DstPort
	}
	if req.Description != "" {
		forward.Description = req.Description
	}
	if req.Enabled != nil {
		forward.Enabled = *req.Enabled
	}

	if result := db.DB.Save(&forward); result.Error != nil {
		return nil, errors.Database("更新转发规则失败", result.Error)
	}

	return &forward, nil
}

// DeleteForward 删除转发规则
func (s *Service) DeleteForward(userID uint, forwardID uint) error {
	var forward db.Forward
	if result := db.DB.Where("id = ? AND user_id = ?", forwardID, userID).First(&forward); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.NotFound("转发规则不存在")
		}
		return errors.Database("查询转发规则失败", result.Error)
	}

	// 删除转发规则
	if result := db.DB.Delete(&forward); result.Error != nil {
		return errors.Database("删除转发规则失败", result.Error)
	}

	return nil
}

// EnableForward 启用转发规则
func (s *Service) EnableForward(userID uint, forwardID uint) (*db.Forward, error) {
	var forward db.Forward
	if result := db.DB.Where("id = ? AND user_id = ?", forwardID, userID).First(&forward); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("转发规则不存在")
		}
		return nil, errors.Database("查询转发规则失败", result.Error)
	}

	// 检查转发规则状态
	if forward.Enabled {
		return nil, errors.Conflict("转发规则已启用")
	}

	// 更新转发规则状态
	forward.Enabled = true
	if result := db.DB.Save(&forward); result.Error != nil {
		return nil, errors.Database("更新转发规则状态失败", result.Error)
	}

	return &forward, nil
}

// DisableForward 禁用转发规则
func (s *Service) DisableForward(userID uint, forwardID uint) (*db.Forward, error) {
	var forward db.Forward
	if result := db.DB.Where("id = ? AND user_id = ?", forwardID, userID).First(&forward); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("转发规则不存在")
		}
		return nil, errors.Database("查询转发规则失败", result.Error)
	}

	// 检查转发规则状态
	if !forward.Enabled {
		return nil, errors.Conflict("转发规则已禁用")
	}

	// 更新转发规则状态
	forward.Enabled = false
	if result := db.DB.Save(&forward); result.Error != nil {
		return nil, errors.Database("更新转发规则状态失败", result.Error)
	}

	return &forward, nil
}

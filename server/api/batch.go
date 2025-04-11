package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/server/db"
)

// BatchHandler 批量操作处理器
type BatchHandler struct {
	db *db.Database
}

// NewBatchHandler 创建批量操作处理器
func NewBatchHandler(db *db.Database) *BatchHandler {
	return &BatchHandler{
		db: db,
	}
}

// BatchDeviceOperation 批量设备操作
func (h *BatchHandler) BatchDeviceOperation(c *gin.Context) {
	var req struct {
		DeviceIDs []uint `json:"deviceIds" binding:"required"`
		Operation string `json:"operation" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 获取当前用户 ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 检查设备权限
	for _, deviceID := range req.DeviceIDs {
		var device db.Device
		if err := h.db.DB.Where("id = ? AND user_id = ?", deviceID, userID).First(&device).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "没有权限操作某些设备"})
			return
		}
	}

	// 执行批量操作
	switch req.Operation {
	case "restart":
		// 重启设备
		for _, deviceID := range req.DeviceIDs {
			// 这里应该调用实际的重启设备的逻辑
			// 为了简化，这里只是更新设备状态
			h.db.DB.Model(&db.Device{}).Where("id = ?", deviceID).Update("status", "restarting")
		}
	case "shutdown":
		// 关闭设备
		for _, deviceID := range req.DeviceIDs {
			// 这里应该调用实际的关闭设备的逻辑
			// 为了简化，这里只是更新设备状态
			h.db.DB.Model(&db.Device{}).Where("id = ?", deviceID).Update("status", "offline")
		}
	case "delete":
		// 删除设备
		for _, deviceID := range req.DeviceIDs {
			h.db.DB.Delete(&db.Device{}, deviceID)
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的操作"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "批量操作成功"})
}

// BatchAppOperation 批量应用操作
func (h *BatchHandler) BatchAppOperation(c *gin.Context) {
	var req struct {
		AppIDs    []uint `json:"appIds" binding:"required"`
		Operation string `json:"operation" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 获取当前用户 ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 检查应用权限
	for _, appID := range req.AppIDs {
		var app db.App
		if err := h.db.DB.Where("id = ?", appID).First(&app).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到某些应用"})
			return
		}

		// 检查设备权限
		var device db.Device
		if err := h.db.DB.Where("id = ? AND user_id = ?", app.DeviceID, userID).First(&device).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "没有权限操作某些应用"})
			return
		}
	}

	// 执行批量操作
	switch req.Operation {
	case "start":
		// 启动应用
		for _, appID := range req.AppIDs {
			// 这里应该调用实际的启动应用的逻辑
			// 为了简化，这里只是更新应用状态
			h.db.DB.Model(&db.App{}).Where("id = ?", appID).Update("status", "running")
		}
	case "stop":
		// 停止应用
		for _, appID := range req.AppIDs {
			// 这里应该调用实际的停止应用的逻辑
			// 为了简化，这里只是更新应用状态
			h.db.DB.Model(&db.App{}).Where("id = ?", appID).Update("status", "stopped")
		}
	case "delete":
		// 删除应用
		for _, appID := range req.AppIDs {
			h.db.DB.Delete(&db.App{}, appID)
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的操作"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "批量操作成功"})
}

// BatchForwardOperation 批量转发规则操作
func (h *BatchHandler) BatchForwardOperation(c *gin.Context) {
	var req struct {
		ForwardIDs []uint `json:"forwardIds" binding:"required"`
		Operation  string `json:"operation" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 获取当前用户 ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 检查转发规则权限
	for _, forwardID := range req.ForwardIDs {
		var forward db.Forward
		if err := h.db.DB.Where("id = ?", forwardID).First(&forward).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到某些转发规则"})
			return
		}

		// 检查用户权限
		if forward.UserID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"error": "没有权限操作某些转发规则"})
			return
		}
	}

	// 执行批量操作
	switch req.Operation {
	case "enable":
		// 启用转发规则
		for _, forwardID := range req.ForwardIDs {
			h.db.DB.Model(&db.Forward{}).Where("id = ?", forwardID).Update("enabled", true)
		}
	case "disable":
		// 禁用转发规则
		for _, forwardID := range req.ForwardIDs {
			h.db.DB.Model(&db.Forward{}).Where("id = ?", forwardID).Update("enabled", false)
		}
	case "delete":
		// 删除转发规则
		for _, forwardID := range req.ForwardIDs {
			h.db.DB.Delete(&db.Forward{}, forwardID)
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的操作"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "批量操作成功"})
}

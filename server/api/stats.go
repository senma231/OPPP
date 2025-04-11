package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/server/db"
)

// StatsHandler 统计处理器
type StatsHandler struct {
	db *db.Database
}

// NewStatsHandler 创建统计处理器
func NewStatsHandler(db *db.Database) *StatsHandler {
	return &StatsHandler{
		db: db,
	}
}

// GetSystemStats 获取系统统计信息
func (h *StatsHandler) GetSystemStats(c *gin.Context) {
	// 获取设备数量
	var deviceCount int64
	if err := h.db.DB.Model(&db.Device{}).Count(&deviceCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取设备数量失败"})
		return
	}

	// 获取在线设备数量
	var onlineDeviceCount int64
	if err := h.db.DB.Model(&db.Device{}).Where("status = ?", "online").Count(&onlineDeviceCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取在线设备数量失败"})
		return
	}

	// 获取应用数量
	var appCount int64
	if err := h.db.DB.Model(&db.App{}).Count(&appCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取应用数量失败"})
		return
	}

	// 获取运行中的应用数量
	var runningAppCount int64
	if err := h.db.DB.Model(&db.App{}).Where("status = ?", "running").Count(&runningAppCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取运行中的应用数量失败"})
		return
	}

	// 获取转发规则数量
	var forwardCount int64
	if err := h.db.DB.Model(&db.Forward{}).Count(&forwardCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取转发规则数量失败"})
		return
	}

	// 获取启用的转发规则数量
	var enabledForwardCount int64
	if err := h.db.DB.Model(&db.Forward{}).Where("enabled = ?", true).Count(&enabledForwardCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取启用的转发规则数量失败"})
		return
	}

	// 获取连接数量
	var connectionCount int64
	if err := h.db.DB.Model(&db.Connection{}).Count(&connectionCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取连接数量失败"})
		return
	}

	// 获取连接类型统计
	var directCount, upnpCount, holePunchCount, relayCount int64
	if err := h.db.DB.Model(&db.Connection{}).Where("type = ?", "direct").Count(&directCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取直接连接数量失败"})
		return
	}
	if err := h.db.DB.Model(&db.Connection{}).Where("type = ?", "upnp").Count(&upnpCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取 UPnP 连接数量失败"})
		return
	}
	if err := h.db.DB.Model(&db.Connection{}).Where("type = ?", "hole_punch").Count(&holePunchCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取打洞连接数量失败"})
		return
	}
	if err := h.db.DB.Model(&db.Connection{}).Where("type = ?", "relay").Count(&relayCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取中继连接数量失败"})
		return
	}

	// 获取流量统计
	var totalSent, totalReceived int64
	h.db.DB.Model(&db.Stats{}).Select("SUM(bytes_sent)").Row().Scan(&totalSent)
	h.db.DB.Model(&db.Stats{}).Select("SUM(bytes_received)").Row().Scan(&totalReceived)

	// 返回统计信息
	c.JSON(http.StatusOK, gin.H{
		"version": "1.0.0",
		"uptime":  int64(time.Since(time.Now().Add(-24 * time.Hour)).Seconds()), // 模拟运行时间
		"devices": gin.H{
			"total":  deviceCount,
			"online": onlineDeviceCount,
		},
		"apps": gin.H{
			"total":   appCount,
			"running": runningAppCount,
		},
		"forwards": gin.H{
			"total":   forwardCount,
			"enabled": enabledForwardCount,
		},
		"connections": gin.H{
			"direct":    directCount,
			"upnp":      upnpCount,
			"holePunch": holePunchCount,
			"relay":     relayCount,
		},
		"traffic": gin.H{
			"sent":     totalSent,
			"received": totalReceived,
		},
	})
}

// GetDeviceStats 获取设备统计信息
func (h *StatsHandler) GetDeviceStats(c *gin.Context) {
	deviceID := c.Param("id")
	
	// 转换设备 ID
	id, err := strconv.ParseUint(deviceID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的设备 ID"})
		return
	}
	
	// 获取设备统计信息
	var stats db.Stats
	if err := h.db.DB.Where("device_id = ?", id).First(&stats).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到设备统计信息"})
		return
	}
	
	// 返回统计信息
	c.JSON(http.StatusOK, gin.H{
		"bytesSent":     stats.BytesSent,
		"bytesReceived": stats.BytesReceived,
		"connections":   stats.Connections,
		"connectionTime": stats.ConnectionTime,
	})
}

// GetAppStats 获取应用统计信息
func (h *StatsHandler) GetAppStats(c *gin.Context) {
	appID := c.Param("id")
	
	// 转换应用 ID
	id, err := strconv.ParseUint(appID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的应用 ID"})
		return
	}
	
	// 获取应用统计信息
	var stats db.Stats
	if err := h.db.DB.Where("app_id = ?", id).First(&stats).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到应用统计信息"})
		return
	}
	
	// 返回统计信息
	c.JSON(http.StatusOK, gin.H{
		"bytesSent":     stats.BytesSent,
		"bytesReceived": stats.BytesReceived,
		"connections":   stats.Connections,
		"connectionTime": stats.ConnectionTime,
	})
}

// GetForwardStats 获取转发规则统计信息
func (h *StatsHandler) GetForwardStats(c *gin.Context) {
	forwardID := c.Param("id")
	
	// 转换转发规则 ID
	id, err := strconv.ParseUint(forwardID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的转发规则 ID"})
		return
	}
	
	// 获取转发规则统计信息
	var stats db.Stats
	if err := h.db.DB.Where("forward_id = ?", id).First(&stats).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到转发规则统计信息"})
		return
	}
	
	// 返回统计信息
	c.JSON(http.StatusOK, gin.H{
		"bytesSent":     stats.BytesSent,
		"bytesReceived": stats.BytesReceived,
		"connections":   stats.Connections,
		"startTime":     stats.CreatedAt,
	})
}

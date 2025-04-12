package api

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/common/errors"
	"github.com/senma231/p3/server/db"
)

// GetSystemStats 获取系统统计信息
func GetSystemStats(c *gin.Context) {
	// 获取用户数量
	var usersCount int64
	if result := db.DB.Model(&db.User{}).Count(&usersCount); result.Error != nil {
		errObj := errors.AsError(result.Error)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	// 获取设备数量
	var devicesCount int64
	if result := db.DB.Model(&db.Device{}).Count(&devicesCount); result.Error != nil {
		errObj := errors.AsError(result.Error)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	// 获取应用数量
	var appsCount int64
	if result := db.DB.Model(&db.App{}).Count(&appsCount); result.Error != nil {
		errObj := errors.AsError(result.Error)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	// 获取在线设备数量
	var onlineDevicesCount int64
	if result := db.DB.Model(&db.Device{}).Where("status = ?", "online").Count(&onlineDevicesCount); result.Error != nil {
		errObj := errors.AsError(result.Error)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	// 获取总连接数
	var totalConnections int64
	if result := db.DB.Model(&db.Stats{}).Sum("connections").Scan(&totalConnections); result.Error != nil {
		errObj := errors.AsError(result.Error)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	// 获取总流量
	var totalBytesSent, totalBytesReceived int64
	if result := db.DB.Model(&db.Stats{}).Sum("bytes_sent").Scan(&totalBytesSent); result.Error != nil {
		errObj := errors.AsError(result.Error)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}
	if result := db.DB.Model(&db.Stats{}).Sum("bytes_received").Scan(&totalBytesReceived); result.Error != nil {
		errObj := errors.AsError(result.Error)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}
	totalTraffic := totalBytesSent + totalBytesReceived

	// 获取系统信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryUsage := float64(m.Alloc) / float64(m.Sys) * 100

	// 获取 CPU 使用率（简化版，实际应该使用系统调用）
	cpuUsage := 0.0

	// 获取系统启动时间
	uptime := time.Since(startTime).Seconds()

	c.JSON(http.StatusOK, gin.H{
		"users_count":      usersCount,
		"devices_count":    devicesCount,
		"apps_count":       appsCount,
		"online_devices":   onlineDevicesCount,
		"total_connections": totalConnections,
		"total_traffic":    totalTraffic,
		"cpu_usage":        cpuUsage,
		"memory_usage":     memoryUsage,
		"uptime":           uptime,
	})
}

// GetUserStats 获取用户统计信息
func GetUserStats(c *gin.Context) {
	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取设备数量
	var devicesCount int64
	if result := db.DB.Model(&db.Device{}).Where("user_id = ?", userID).Count(&devicesCount); result.Error != nil {
		errObj := errors.AsError(result.Error)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	// 获取应用数量
	var appsCount int64
	if result := db.DB.Model(&db.App{}).Where("user_id = ?", userID).Count(&appsCount); result.Error != nil {
		errObj := errors.AsError(result.Error)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	// 获取在线设备数量
	var onlineDevicesCount int64
	if result := db.DB.Model(&db.Device{}).Where("user_id = ? AND status = ?", userID, "online").Count(&onlineDevicesCount); result.Error != nil {
		errObj := errors.AsError(result.Error)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	// 获取总连接数
	var totalConnections int64
	if result := db.DB.Model(&db.Stats{}).Where("user_id = ?", userID).Sum("connections").Scan(&totalConnections); result.Error != nil {
		errObj := errors.AsError(result.Error)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	// 获取总流量
	var totalBytesSent, totalBytesReceived int64
	if result := db.DB.Model(&db.Stats{}).Where("user_id = ?", userID).Sum("bytes_sent").Scan(&totalBytesSent); result.Error != nil {
		errObj := errors.AsError(result.Error)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}
	if result := db.DB.Model(&db.Stats{}).Where("user_id = ?", userID).Sum("bytes_received").Scan(&totalBytesReceived); result.Error != nil {
		errObj := errors.AsError(result.Error)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}
	totalTraffic := totalBytesSent + totalBytesReceived

	// 获取活跃连接数
	var activeConnections int64
	if result := db.DB.Model(&db.App{}).Where("user_id = ? AND status = ?", userID, "running").Count(&activeConnections); result.Error != nil {
		errObj := errors.AsError(result.Error)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"devices_count":     devicesCount,
		"apps_count":        appsCount,
		"online_devices":    onlineDevicesCount,
		"total_connections": totalConnections,
		"total_traffic":     totalTraffic,
		"active_connections": activeConnections,
	})
}

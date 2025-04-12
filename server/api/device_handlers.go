package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/common/errors"
	"github.com/senma231/p3/server/device"
)

// GetDevices 获取设备列表
func GetDevices(c *gin.Context) {
	// 获取设备服务
	deviceService := c.MustGet("deviceService").(*device.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取设备列表
	devices, err := deviceService.GetDevices(userID)
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"devices": devices,
	})
}

// GetDevice 获取设备详情
func GetDevice(c *gin.Context) {
	// 获取设备服务
	deviceService := c.MustGet("deviceService").(*device.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取设备 ID
	deviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的设备 ID",
		})
		return
	}

	// 获取设备详情
	device, err := deviceService.GetDevice(userID, uint(deviceID))
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, device)
}

// CreateDevice 创建设备
func CreateDevice(c *gin.Context) {
	var req device.DeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
		})
		return
	}

	// 获取设备服务
	deviceService := c.MustGet("deviceService").(*device.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 创建设备
	device, err := deviceService.CreateDevice(userID, &req)
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, device)
}

// UpdateDevice 更新设备
func UpdateDevice(c *gin.Context) {
	var req device.DeviceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
		})
		return
	}

	// 获取设备服务
	deviceService := c.MustGet("deviceService").(*device.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取设备 ID
	deviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的设备 ID",
		})
		return
	}

	// 更新设备
	device, err := deviceService.UpdateDevice(userID, uint(deviceID), &req)
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, device)
}

// DeleteDevice 删除设备
func DeleteDevice(c *gin.Context) {
	// 获取设备服务
	deviceService := c.MustGet("deviceService").(*device.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取设备 ID
	deviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的设备 ID",
		})
		return
	}

	// 删除设备
	if err := deviceService.DeleteDevice(userID, uint(deviceID)); err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设备已成功删除",
	})
}

// RegenerateDeviceToken 重新生成设备令牌
func RegenerateDeviceToken(c *gin.Context) {
	// 获取设备服务
	deviceService := c.MustGet("deviceService").(*device.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取设备 ID
	deviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的设备 ID",
		})
		return
	}

	// 重新生成设备令牌
	token, err := deviceService.RegenerateToken(userID, uint(deviceID))
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// UpdateDeviceStatus 更新设备状态
func UpdateDeviceStatus(c *gin.Context) {
	var req device.DeviceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
		})
		return
	}

	// 获取设备服务
	deviceService := c.MustGet("deviceService").(*device.Service)

	// 从上下文中获取设备 ID
	deviceID := c.MustGet("deviceID").(uint)

	// 更新设备状态
	device, err := deviceService.UpdateDeviceStatus(deviceID, &req)
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, device)
}

// GetDeviceApps 获取设备应用列表
func GetDeviceApps(c *gin.Context) {
	// 获取应用服务
	appService := c.MustGet("appService").(*app.Service)

	// 从上下文中获取设备 ID
	deviceID := c.MustGet("deviceID").(uint)

	// 获取设备应用列表
	apps, err := appService.GetAppsByDevice(deviceID)
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"apps": apps,
	})
}

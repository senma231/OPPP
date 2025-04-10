package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/server/device"
)

// DeviceController 设备控制器
type DeviceController struct {
	deviceService *device.Service
}

// NewDeviceController 创建设备控制器
func NewDeviceController(deviceService *device.Service) *DeviceController {
	return &DeviceController{
		deviceService: deviceService,
	}
}

// GetDevices 获取设备列表
func (c *DeviceController) GetDevices(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	devices, err := c.deviceService.GetDevicesByUserID(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"devices": devices,
	})
}

// GetDevice 获取设备详情
func (c *DeviceController) GetDevice(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	deviceID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的设备 ID",
		})
		return
	}

	device, err := c.deviceService.GetDeviceByID(uint(deviceID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 检查设备是否属于当前用户
	if device.UserID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "无权访问该设备",
		})
		return
	}

	ctx.JSON(http.StatusOK, device)
}

// CreateDevice 创建设备
func (c *DeviceController) CreateDevice(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	var req struct {
		Name   string `json:"name" binding:"required"`
		NodeID string `json:"nodeId" binding:"required"`
		Token  string `json:"token" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
		})
		return
	}

	device, err := c.deviceService.CreateDevice(userID.(uint), req.Name, req.NodeID, req.Token)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, device)
}

// UpdateDevice 更新设备
func (c *DeviceController) UpdateDevice(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	deviceID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的设备 ID",
		})
		return
	}

	device, err := c.deviceService.GetDeviceByID(uint(deviceID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 检查设备是否属于当前用户
	if device.UserID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "无权修改该设备",
		})
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
		})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}

	updatedDevice, err := c.deviceService.UpdateDevice(uint(deviceID), updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, updatedDevice)
}

// DeleteDevice 删除设备
func (c *DeviceController) DeleteDevice(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	deviceID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的设备 ID",
		})
		return
	}

	device, err := c.deviceService.GetDeviceByID(uint(deviceID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 检查设备是否属于当前用户
	if device.UserID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "无权删除该设备",
		})
		return
	}

	if err := c.deviceService.DeleteDevice(uint(deviceID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "设备已删除",
	})
}

// GetDeviceStats 获取设备统计信息
func (c *DeviceController) GetDeviceStats(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	deviceID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的设备 ID",
		})
		return
	}

	device, err := c.deviceService.GetDeviceByID(uint(deviceID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 检查设备是否属于当前用户
	if device.UserID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "无权访问该设备",
		})
		return
	}

	stats, err := c.deviceService.GetDeviceStats(uint(deviceID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

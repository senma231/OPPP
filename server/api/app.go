package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/server/app"
)

// AppController 应用控制器
type AppController struct {
	appService *app.Service
}

// NewAppController 创建应用控制器
func NewAppController(appService *app.Service) *AppController {
	return &AppController{
		appService: appService,
	}
}

// GetApps 获取应用列表
func (c *AppController) GetApps(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	deviceID := ctx.Query("deviceId")
	if deviceID != "" {
		// 获取指定设备的应用
		id, err := strconv.ParseUint(deviceID, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的设备 ID",
			})
			return
		}

		apps, err := c.appService.GetAppsByDeviceID(uint(id))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"apps": apps,
		})
		return
	}

	// 获取用户的所有应用
	apps, err := c.appService.GetAppsByUserID(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"apps": apps,
	})
}

// GetApp 获取应用详情
func (c *AppController) GetApp(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	appID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的应用 ID",
		})
		return
	}

	app, err := c.appService.GetAppByID(uint(appID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 检查应用是否属于当前用户
	if app.UserID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "无权访问该应用",
		})
		return
	}

	ctx.JSON(http.StatusOK, app)
}

// CreateApp 创建应用
func (c *AppController) CreateApp(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	var req struct {
		DeviceID    uint   `json:"deviceId" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Protocol    string `json:"protocol" binding:"required"`
		SrcPort     int    `json:"srcPort" binding:"required"`
		PeerNode    string `json:"peerNode" binding:"required"`
		DstPort     int    `json:"dstPort" binding:"required"`
		DstHost     string `json:"dstHost" binding:"required"`
		Description string `json:"description"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
		})
		return
	}

	app, err := c.appService.CreateApp(
		userID.(uint),
		req.DeviceID,
		req.Name,
		req.Protocol,
		req.SrcPort,
		req.PeerNode,
		req.DstPort,
		req.DstHost,
		req.Description,
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, app)
}

// UpdateApp 更新应用
func (c *AppController) UpdateApp(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	appID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的应用 ID",
		})
		return
	}

	app, err := c.appService.GetAppByID(uint(appID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 检查应用是否属于当前用户
	if app.UserID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "无权修改该应用",
		})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Protocol    string `json:"protocol"`
		SrcPort     int    `json:"srcPort"`
		PeerNode    string `json:"peerNode"`
		DstPort     int    `json:"dstPort"`
		DstHost     string `json:"dstHost"`
		Description string `json:"description"`
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
	if req.Protocol != "" {
		updates["protocol"] = req.Protocol
	}
	if req.SrcPort != 0 {
		updates["src_port"] = req.SrcPort
	}
	if req.PeerNode != "" {
		updates["peer_node"] = req.PeerNode
	}
	if req.DstPort != 0 {
		updates["dst_port"] = req.DstPort
	}
	if req.DstHost != "" {
		updates["dst_host"] = req.DstHost
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	updatedApp, err := c.appService.UpdateApp(uint(appID), updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, updatedApp)
}

// DeleteApp 删除应用
func (c *AppController) DeleteApp(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	appID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的应用 ID",
		})
		return
	}

	app, err := c.appService.GetAppByID(uint(appID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 检查应用是否属于当前用户
	if app.UserID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "无权删除该应用",
		})
		return
	}

	if err := c.appService.DeleteApp(uint(appID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "应用已删除",
	})
}

// StartApp 启动应用
func (c *AppController) StartApp(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	appID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的应用 ID",
		})
		return
	}

	app, err := c.appService.GetAppByID(uint(appID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 检查应用是否属于当前用户
	if app.UserID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "无权操作该应用",
		})
		return
	}

	app, err = c.appService.StartApp(uint(appID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":     app.ID,
		"name":   app.Name,
		"status": app.Status,
	})
}

// StopApp 停止应用
func (c *AppController) StopApp(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	appID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的应用 ID",
		})
		return
	}

	app, err := c.appService.GetAppByID(uint(appID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 检查应用是否属于当前用户
	if app.UserID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "无权操作该应用",
		})
		return
	}

	app, err = c.appService.StopApp(uint(appID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":     app.ID,
		"name":   app.Name,
		"status": app.Status,
	})
}

// GetAppStats 获取应用统计信息
func (c *AppController) GetAppStats(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	appID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的应用 ID",
		})
		return
	}

	app, err := c.appService.GetAppByID(uint(appID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 检查应用是否属于当前用户
	if app.UserID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "无权访问该应用",
		})
		return
	}

	stats, err := c.appService.GetAppStats(uint(appID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

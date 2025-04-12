package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/common/errors"
	"github.com/senma231/p3/server/app"
)

// GetApps 获取应用列表
func GetApps(c *gin.Context) {
	// 获取应用服务
	appService := c.MustGet("appService").(*app.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取应用列表
	apps, err := appService.GetApps(userID)
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

// GetApp 获取应用详情
func GetApp(c *gin.Context) {
	// 获取应用服务
	appService := c.MustGet("appService").(*app.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取应用 ID
	appID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的应用 ID",
		})
		return
	}

	// 获取应用详情
	app, err := appService.GetApp(userID, uint(appID))
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, app)
}

// CreateApp 创建应用
func CreateApp(c *gin.Context) {
	var req app.AppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
		})
		return
	}

	// 获取应用服务
	appService := c.MustGet("appService").(*app.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取设备 ID
	deviceID, err := strconv.ParseUint(c.Query("device_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的设备 ID",
		})
		return
	}

	// 创建应用
	app, err := appService.CreateApp(userID, uint(deviceID), &req)
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, app)
}

// UpdateApp 更新应用
func UpdateApp(c *gin.Context) {
	var req app.AppUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
		})
		return
	}

	// 获取应用服务
	appService := c.MustGet("appService").(*app.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取应用 ID
	appID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的应用 ID",
		})
		return
	}

	// 更新应用
	app, err := appService.UpdateApp(userID, uint(appID), &req)
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, app)
}

// DeleteApp 删除应用
func DeleteApp(c *gin.Context) {
	// 获取应用服务
	appService := c.MustGet("appService").(*app.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取应用 ID
	appID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的应用 ID",
		})
		return
	}

	// 删除应用
	if err := appService.DeleteApp(userID, uint(appID)); err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "应用已成功删除",
	})
}

// StartApp 启动应用
func StartApp(c *gin.Context) {
	// 获取应用服务
	appService := c.MustGet("appService").(*app.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取应用 ID
	appID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的应用 ID",
		})
		return
	}

	// 启动应用
	app, err := appService.StartApp(userID, uint(appID))
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, app)
}

// StopApp 停止应用
func StopApp(c *gin.Context) {
	// 获取应用服务
	appService := c.MustGet("appService").(*app.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取应用 ID
	appID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的应用 ID",
		})
		return
	}

	// 停止应用
	app, err := appService.StopApp(userID, uint(appID))
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, app)
}

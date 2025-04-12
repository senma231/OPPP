package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/common/errors"
	"github.com/senma231/p3/server/forward"
)

// GetForwards 获取转发规则列表
func GetForwards(c *gin.Context) {
	// 获取转发服务
	forwardService := c.MustGet("forwardService").(*forward.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取转发规则列表
	forwards, err := forwardService.GetForwards(userID)
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"forwards": forwards,
	})
}

// GetForward 获取转发规则详情
func GetForward(c *gin.Context) {
	// 获取转发服务
	forwardService := c.MustGet("forwardService").(*forward.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取转发规则 ID
	forwardID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的转发规则 ID",
		})
		return
	}

	// 获取转发规则详情
	forward, err := forwardService.GetForward(userID, uint(forwardID))
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, forward)
}

// CreateForward 创建转发规则
func CreateForward(c *gin.Context) {
	var req forward.ForwardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
		})
		return
	}

	// 获取转发服务
	forwardService := c.MustGet("forwardService").(*forward.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 创建转发规则
	forward, err := forwardService.CreateForward(userID, &req)
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, forward)
}

// UpdateForward 更新转发规则
func UpdateForward(c *gin.Context) {
	var req forward.ForwardUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
		})
		return
	}

	// 获取转发服务
	forwardService := c.MustGet("forwardService").(*forward.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取转发规则 ID
	forwardID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的转发规则 ID",
		})
		return
	}

	// 更新转发规则
	forward, err := forwardService.UpdateForward(userID, uint(forwardID), &req)
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, forward)
}

// DeleteForward 删除转发规则
func DeleteForward(c *gin.Context) {
	// 获取转发服务
	forwardService := c.MustGet("forwardService").(*forward.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取转发规则 ID
	forwardID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的转发规则 ID",
		})
		return
	}

	// 删除转发规则
	if err := forwardService.DeleteForward(userID, uint(forwardID)); err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "转发规则已成功删除",
	})
}

// EnableForward 启用转发规则
func EnableForward(c *gin.Context) {
	// 获取转发服务
	forwardService := c.MustGet("forwardService").(*forward.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取转发规则 ID
	forwardID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的转发规则 ID",
		})
		return
	}

	// 启用转发规则
	forward, err := forwardService.EnableForward(userID, uint(forwardID))
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, forward)
}

// DisableForward 禁用转发规则
func DisableForward(c *gin.Context) {
	// 获取转发服务
	forwardService := c.MustGet("forwardService").(*forward.Service)

	// 从上下文中获取用户 ID
	userID := c.MustGet("userID").(uint)

	// 获取转发规则 ID
	forwardID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的转发规则 ID",
		})
		return
	}

	// 禁用转发规则
	forward, err := forwardService.DisableForward(userID, uint(forwardID))
	if err != nil {
		errObj := errors.AsError(err)
		c.JSON(errObj.StatusCode(), gin.H{
			"error": errObj.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, forward)
}

package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/server/auth"
	"github.com/senma231/p3/server/db"
)

// PermissionHandler 权限处理器
type PermissionHandler struct {
	db        *db.Database
	permMgr   *auth.PermissionManager
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(db *db.Database, permMgr *auth.PermissionManager) *PermissionHandler {
	return &PermissionHandler{
		db:      db,
		permMgr: permMgr,
	}
}

// GetUserRole 获取用户角色
func (h *PermissionHandler) GetUserRole(c *gin.Context) {
	userID := c.Param("id")

	// 转换用户 ID
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
		return
	}

	// 获取当前用户 ID
	currentUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 检查权限
	hasPermission, err := h.permMgr.HasPermission(currentUserID.(uint), auth.PermissionReadUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}
	if !hasPermission && currentUserID.(uint) != uint(id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限查看其他用户的角色"})
		return
	}

	// 获取用户角色
	role, err := h.permMgr.GetUserRole(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到用户"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"role": role})
}

// SetUserRole 设置用户角色
func (h *PermissionHandler) SetUserRole(c *gin.Context) {
	userID := c.Param("id")

	// 转换用户 ID
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
		return
	}

	// 获取当前用户 ID
	currentUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 检查权限
	hasPermission, err := h.permMgr.HasPermission(currentUserID.(uint), auth.PermissionWriteUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限修改用户角色"})
		return
	}

	// 绑定请求数据
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 验证角色
	role := auth.Role(req.Role)
	if role != auth.RoleAdmin && role != auth.RoleUser && role != auth.RoleGuest {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色"})
		return
	}

	// 设置用户角色
	if err := h.permMgr.SetUserRole(uint(id), role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置用户角色失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户角色已更新"})
}

// GetUserPermissions 获取用户权限
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	userID := c.Param("id")

	// 转换用户 ID
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
		return
	}

	// 获取当前用户 ID
	currentUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 检查权限
	hasPermission, err := h.permMgr.HasPermission(currentUserID.(uint), auth.PermissionReadUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}
	if !hasPermission && currentUserID.(uint) != uint(id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限查看其他用户的权限"})
		return
	}

	// 获取用户权限
	permissions, err := h.permMgr.GetUserPermissions(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到用户"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// CheckPermission 检查权限
func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	// 绑定请求数据
	var req struct {
		Permission  string `json:"permission" binding:"required"`
		ResourceType string `json:"resourceType"`
		ResourceID   uint   `json:"resourceId"`
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

	// 检查权限
	permission := auth.Permission(req.Permission)
	var hasPermission bool
	var err error

	if req.ResourceType != "" && req.ResourceID != 0 {
		// 检查资源权限
		hasPermission, err = h.permMgr.HasResourcePermission(userID.(uint), req.ResourceType, req.ResourceID, permission)
	} else {
		// 检查普通权限
		hasPermission, err = h.permMgr.HasPermission(userID.(uint), permission)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查权限失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"hasPermission": hasPermission})
}

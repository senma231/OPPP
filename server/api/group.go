package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/server/db"
)

// GroupHandler 分组处理器
type GroupHandler struct {
	db *db.Database
}

// NewGroupHandler 创建分组处理器
func NewGroupHandler(db *db.Database) *GroupHandler {
	return &GroupHandler{
		db: db,
	}
}

// CreateGroup 创建分组
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var group db.Group
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 获取当前用户 ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	group.UserID = userID.(uint)

	// 创建分组
	if err := h.db.CreateGroup(&group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建分组失败"})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// GetGroups 获取分组列表
func (h *GroupHandler) GetGroups(c *gin.Context) {
	// 获取当前用户 ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取用户的所有分组
	groups, err := h.db.GetGroupsByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取分组列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

// GetGroup 获取分组详情
func (h *GroupHandler) GetGroup(c *gin.Context) {
	groupID := c.Param("id")

	// 转换分组 ID
	id, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分组 ID"})
		return
	}

	// 获取分组详情
	group, err := h.db.GetGroupByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到分组"})
		return
	}

	// 检查权限
	userID, exists := c.Get("userID")
	if !exists || group.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限访问此分组"})
		return
	}

	c.JSON(http.StatusOK, group)
}

// UpdateGroup 更新分组
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	groupID := c.Param("id")

	// 转换分组 ID
	id, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分组 ID"})
		return
	}

	// 获取分组详情
	group, err := h.db.GetGroupByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到分组"})
		return
	}

	// 检查权限
	userID, exists := c.Get("userID")
	if !exists || group.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限更新此分组"})
		return
	}

	// 绑定请求数据
	var updateData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 更新分组
	group.Name = updateData.Name
	group.Description = updateData.Description
	if err := h.db.UpdateGroup(group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新分组失败"})
		return
	}

	c.JSON(http.StatusOK, group)
}

// DeleteGroup 删除分组
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	groupID := c.Param("id")

	// 转换分组 ID
	id, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分组 ID"})
		return
	}

	// 获取分组详情
	group, err := h.db.GetGroupByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到分组"})
		return
	}

	// 检查权限
	userID, exists := c.Get("userID")
	if !exists || group.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限删除此分组"})
		return
	}

	// 删除分组
	if err := h.db.DeleteGroup(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除分组失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "分组已删除"})
}

// AddDeviceToGroup 添加设备到分组
func (h *GroupHandler) AddDeviceToGroup(c *gin.Context) {
	groupID := c.Param("id")
	deviceID := c.Param("deviceId")

	// 转换 ID
	gid, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分组 ID"})
		return
	}
	did, err := strconv.ParseUint(deviceID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的设备 ID"})
		return
	}

	// 获取分组详情
	group, err := h.db.GetGroupByID(uint(gid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到分组"})
		return
	}

	// 检查权限
	userID, exists := c.Get("userID")
	if !exists || group.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限修改此分组"})
		return
	}

	// 添加设备到分组
	if err := h.db.AddDeviceToGroup(uint(gid), uint(did)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加设备到分组失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "设备已添加到分组"})
}

// RemoveDeviceFromGroup 从分组中移除设备
func (h *GroupHandler) RemoveDeviceFromGroup(c *gin.Context) {
	groupID := c.Param("id")
	deviceID := c.Param("deviceId")

	// 转换 ID
	gid, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分组 ID"})
		return
	}
	did, err := strconv.ParseUint(deviceID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的设备 ID"})
		return
	}

	// 获取分组详情
	group, err := h.db.GetGroupByID(uint(gid))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到分组"})
		return
	}

	// 检查权限
	userID, exists := c.Get("userID")
	if !exists || group.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限修改此分组"})
		return
	}

	// 从分组中移除设备
	if err := h.db.RemoveDeviceFromGroup(uint(gid), uint(did)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "从分组中移除设备失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "设备已从分组中移除"})
}

// GetDevicesInGroup 获取分组中的设备
func (h *GroupHandler) GetDevicesInGroup(c *gin.Context) {
	groupID := c.Param("id")

	// 转换分组 ID
	id, err := strconv.ParseUint(groupID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分组 ID"})
		return
	}

	// 获取分组详情
	group, err := h.db.GetGroupByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到分组"})
		return
	}

	// 检查权限
	userID, exists := c.Get("userID")
	if !exists || group.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限访问此分组"})
		return
	}

	// 获取分组中的设备
	devices, err := h.db.GetDevicesByGroupID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取分组中的设备失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"devices": devices})
}

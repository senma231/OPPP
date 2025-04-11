package auth

import (
	"errors"
	"time"

	"github.com/senma231/p3/server/db"
)

// Role 角色
type Role string

const (
	// RoleAdmin 管理员
	RoleAdmin Role = "admin"
	// RoleUser 普通用户
	RoleUser Role = "user"
	// RoleGuest 访客
	RoleGuest Role = "guest"
)

// Permission 权限
type Permission string

const (
	// PermissionReadDevice 读取设备
	PermissionReadDevice Permission = "read:device"
	// PermissionWriteDevice 写入设备
	PermissionWriteDevice Permission = "write:device"
	// PermissionReadApp 读取应用
	PermissionReadApp Permission = "read:app"
	// PermissionWriteApp 写入应用
	PermissionWriteApp Permission = "write:app"
	// PermissionReadForward 读取转发规则
	PermissionReadForward Permission = "read:forward"
	// PermissionWriteForward 写入转发规则
	PermissionWriteForward Permission = "write:forward"
	// PermissionReadUser 读取用户
	PermissionReadUser Permission = "read:user"
	// PermissionWriteUser 写入用户
	PermissionWriteUser Permission = "write:user"
)

// RolePermissions 角色权限映射
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermissionReadDevice, PermissionWriteDevice,
		PermissionReadApp, PermissionWriteApp,
		PermissionReadForward, PermissionWriteForward,
		PermissionReadUser, PermissionWriteUser,
	},
	RoleUser: {
		PermissionReadDevice, PermissionWriteDevice,
		PermissionReadApp, PermissionWriteApp,
		PermissionReadForward, PermissionWriteForward,
	},
	RoleGuest: {
		PermissionReadDevice,
		PermissionReadApp,
		PermissionReadForward,
	},
}

// PermissionManager 权限管理器
type PermissionManager struct {
	db *db.Database
}

// NewPermissionManager 创建权限管理器
func NewPermissionManager(db *db.Database) *PermissionManager {
	return &PermissionManager{
		db: db,
	}
}

// HasPermission 检查用户是否有权限
func (m *PermissionManager) HasPermission(userID uint, permission Permission) (bool, error) {
	// 获取用户
	var user db.User
	if err := m.db.DB.First(&user, userID).Error; err != nil {
		return false, err
	}

	// 获取用户角色
	role := Role(user.Role)

	// 检查角色是否有权限
	permissions, exists := RolePermissions[role]
	if !exists {
		return false, errors.New("未知的角色")
	}

	// 检查权限
	for _, p := range permissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

// HasResourcePermission 检查用户是否有资源权限
func (m *PermissionManager) HasResourcePermission(userID uint, resourceType string, resourceID uint, permission Permission) (bool, error) {
	// 先检查用户是否有权限
	hasPermission, err := m.HasPermission(userID, permission)
	if err != nil {
		return false, err
	}
	if !hasPermission {
		return false, nil
	}

	// 检查资源所有权
	switch resourceType {
	case "device":
		var device db.Device
		if err := m.db.DB.First(&device, resourceID).Error; err != nil {
			return false, err
		}
		return device.UserID == userID, nil
	case "app":
		var app db.App
		if err := m.db.DB.First(&app, resourceID).Error; err != nil {
			return false, err
		}
		// 检查应用所属设备的所有权
		var device db.Device
		if err := m.db.DB.First(&device, app.DeviceID).Error; err != nil {
			return false, err
		}
		return device.UserID == userID, nil
	case "forward":
		var forward db.Forward
		if err := m.db.DB.First(&forward, resourceID).Error; err != nil {
			return false, err
		}
		return forward.UserID == userID, nil
	case "user":
		// 只有管理员可以操作其他用户
		var user db.User
		if err := m.db.DB.First(&user, userID).Error; err != nil {
			return false, err
		}
		return user.Role == string(RoleAdmin) || userID == resourceID, nil
	default:
		return false, errors.New("未知的资源类型")
	}
}

// SetUserRole 设置用户角色
func (m *PermissionManager) SetUserRole(userID uint, role Role) error {
	return m.db.DB.Model(&db.User{}).Where("id = ?", userID).Update("role", string(role)).Error
}

// GetUserRole 获取用户角色
func (m *PermissionManager) GetUserRole(userID uint) (Role, error) {
	var user db.User
	if err := m.db.DB.First(&user, userID).Error; err != nil {
		return "", err
	}
	return Role(user.Role), nil
}

// GetUserPermissions 获取用户权限
func (m *PermissionManager) GetUserPermissions(userID uint) ([]Permission, error) {
	// 获取用户角色
	role, err := m.GetUserRole(userID)
	if err != nil {
		return nil, err
	}

	// 获取角色权限
	permissions, exists := RolePermissions[role]
	if !exists {
		return nil, errors.New("未知的角色")
	}

	return permissions, nil
}

package db

import (
	"time"
)

// Group 设备分组
type Group struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	UserID      uint      `gorm:"not null" json:"userId"`
	Devices     []Device  `gorm:"many2many:group_devices;" json:"devices,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// GroupDevice 分组设备关联
type GroupDevice struct {
	GroupID   uint      `gorm:"primaryKey" json:"groupId"`
	DeviceID  uint      `gorm:"primaryKey" json:"deviceId"`
	CreatedAt time.Time `json:"createdAt"`
}

// CreateGroup 创建分组
func (db *Database) CreateGroup(group *Group) error {
	return db.DB.Create(group).Error
}

// GetGroupByID 根据 ID 获取分组
func (db *Database) GetGroupByID(id uint) (*Group, error) {
	var group Group
	err := db.DB.Preload("Devices").First(&group, id).Error
	return &group, err
}

// GetGroupsByUserID 获取用户的所有分组
func (db *Database) GetGroupsByUserID(userID uint) ([]Group, error) {
	var groups []Group
	err := db.DB.Where("user_id = ?", userID).Find(&groups).Error
	return groups, err
}

// UpdateGroup 更新分组
func (db *Database) UpdateGroup(group *Group) error {
	return db.DB.Save(group).Error
}

// DeleteGroup 删除分组
func (db *Database) DeleteGroup(id uint) error {
	return db.DB.Delete(&Group{}, id).Error
}

// AddDeviceToGroup 添加设备到分组
func (db *Database) AddDeviceToGroup(groupID, deviceID uint) error {
	return db.DB.Create(&GroupDevice{
		GroupID:   groupID,
		DeviceID:  deviceID,
		CreatedAt: time.Now(),
	}).Error
}

// RemoveDeviceFromGroup 从分组中移除设备
func (db *Database) RemoveDeviceFromGroup(groupID, deviceID uint) error {
	return db.DB.Where("group_id = ? AND device_id = ?", groupID, deviceID).Delete(&GroupDevice{}).Error
}

// GetDevicesByGroupID 获取分组中的所有设备
func (db *Database) GetDevicesByGroupID(groupID uint) ([]Device, error) {
	var devices []Device
	err := db.DB.Joins("JOIN group_devices ON group_devices.device_id = devices.id").
		Where("group_devices.group_id = ?", groupID).
		Find(&devices).Error
	return devices, err
}

// GetGroupsByDeviceID 获取设备所属的所有分组
func (db *Database) GetGroupsByDeviceID(deviceID uint) ([]Group, error) {
	var groups []Group
	err := db.DB.Joins("JOIN group_devices ON group_devices.group_id = groups.id").
		Where("group_devices.device_id = ?", deviceID).
		Find(&groups).Error
	return groups, err
}

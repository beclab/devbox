package model

import "time"

type DevContainers struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	DevEnv     string    `gorm:"type:varchar(256);not null;column:dev_env" json:"devEnv"`
	Name       string    `gorm:"type:varchar(256);not null;column:name" json:"devContainerName"`
	CreateTime time.Time `gorm:"default:CURRENT_TIMESTAMP;column:create_time" json:"createTime"`
	UpdateTime time.Time `gorm:"default:CURRENT_TIMESTAMP;column:update_time" json:"updateTime"`
}

func (dc DevContainers) TableName() string {
	return "dev_containers"
}

type DevContainerInfo struct {
	DevContainers
	PodSelector   *string `json:"podSelector,omitempty"`
	AppID         *int    `json:"appId,omitempty"`
	ContainerName *string `json:"containerName,omitempty"`
	Image         *string `json:"image,omitempty"`
	AppName       *string `json:"appName,omitempty"`
	State         *string `json:"state,omitempty"`
	DevPath       *string `json:"devPath,omitempty"`
	Icon          *string `json:"icon"`
}

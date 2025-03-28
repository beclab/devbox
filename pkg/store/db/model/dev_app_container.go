package model

import "time"

type DevAppContainers struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	AppID         uint      `gorm:"column:app_id" json:"appId"`
	ContainerID   uint      `gorm:"column:container_id" json:"containerId"`
	PodSelector   string    `gorm:"type:varchar(128);column:pod_selector" json:"podSelector"`
	ContainerName string    `gorm:"type:varchar(50);column:container_name" json:"containerName"`
	Image         string    `gorm:"type:varchar(128);column:image" json:"image"`
	CreateTime    time.Time `gorm:"default:CURRENT_TIMESTAMP;column:create_time" json:"createTime"`
	UpdateTime    time.Time `gorm:"default:CURRENT_TIMESTAMP;column:update_time" json:"updateTime"`

	Container *DevContainers `gorm:"-" json:"container"`
}

func (dac DevAppContainers) TableName() string {
	return "dev_app_containers"
}

package model

import "time"

type DevApp struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	AppName     string    `gorm:"type:varchar(50);not null;column:app_name;index:app_name" json:"appName"`
	DevEnv      string    `gorm:"type:varchar(50);not null;column:dev_env" json:"devEnv"`
	AppType     string    `gorm:"type:varchar(20);column:app_type" json:"appType"`
	Description string    `gorm:"type:text;column:description" json:"description"`
	CreateTime  time.Time `gorm:"default:CURRENT_TIMESTAMP;column:create_time" json:"createTime"`
	UpdateTime  time.Time `gorm:"default:CURRENT_TIMESTAMP;column:update_time;index:update_time" json:"updateTime"`

	AppID         string                         `gorm:"-" json:"appID"`
	Chart         string                         `gorm:"-" json:"chart"`
	Entrance      string                         `gorm:"-" json:"entrance"`
	PodContainers map[string][]*DevAppContainers `gorm:"-" json:"podContainers"`
}

func (da DevApp) TableName() string {
	return "dev_apps"
}

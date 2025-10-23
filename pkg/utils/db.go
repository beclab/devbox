package utils

import (
	"github.com/beclab/devbox/pkg/store/db"
	"github.com/beclab/devbox/pkg/store/db/model"
	"k8s.io/klog/v2"
)

func UpdateDevApp(owner, name string, updates map[string]interface{}) (appId int64, err error) {
	op := db.NewDbOperator()
	var exists *model.DevApp
	err = op.DB.Where("owner = ?", owner).Where("app_name = ?", name).First(&exists).Error
	if err != nil {
		return 0, err
	}

	err = op.DB.Model(&exists).Updates(updates).Error
	if err != nil {
		klog.Errorf("update dev_app err %v", err)
		return 0, err
	}
	appId = int64(exists.ID)
	return appId, nil
}

package command

import (
	"github.com/beclab/devbox/pkg/utils"
)

type CreateDevContainerConfig struct {
	DevEnv         string `json:"devEnv"`
	Title          string `json:"title"`
	RequiredCpu    string `json:"requiredCpu"`
	RequiredMemory string `json:"requiredMemory"`
	RequiredDisk   string `json:"requiredDisk"`
}

var createConfigDev = &CreateWithOneDockerConfig{
	Container: CreateWithOneDockerContainer{
		Image: "beclab/studio-app:0.0.1",
		Port:  80,
	},
	RequiredCpu:    "50m",
	RequiredMemory: "100Mi",
	RequiredDisk:   "50Mi",
	RequiredGpu:    false,
	NeedPg:         false,
	NeedRedis:      false,
}

func CreateAppWithDevConfig(cfg *CreateDevContainerConfig, owner, name string) error {
	appPath := utils.GetAppPath(owner, name)

	createConfigDev.Name = name
	createConfigDev.Title = cfg.Title
	if cfg != nil {
		if cfg.RequiredCpu != "" {
			createConfigDev.RequiredCpu = cfg.RequiredCpu
		}
		if cfg.RequiredMemory != "" {
			createConfigDev.RequiredMemory = cfg.RequiredMemory
		}
		if cfg.RequiredDisk != "" {
			createConfigDev.RequiredDisk = cfg.RequiredDisk
		}
	}
	at := AppTemplate{}
	at.WithDockerCfg(createConfigDev).WithDockerDeployment(createConfigDev).
		WithDockerService(createConfigDev).WithDockerChartMetadata(createConfigDev).WithDockerOwner(createConfigDev)
	err := at.WriteDockerFile(createConfigDev, appPath)
	if err != nil {
		return err
	}
	return nil
}

package command

import (
	"github.com/beclab/devbox/pkg/utils"
)

type CreateDevContainerConfig struct {
	DevEnv         string `json:"devEnv" validate:"required,devEnv"`
	Title          string `json:"title"`
	RequiredCpu    string `json:"requiredCpu"`
	RequiredMemory string `json:"requiredMemory"`
	RequiredDisk   string `json:"requiredDisk"`
	RequiredGpu    bool   `json:"requiredGpu"`
	ExposePorts    string `json:"exposePorts"`
	GpuVendor      string `json:"gpuVendor"`
	SshEnable      bool   `json:"sshEnable"`
}

var createConfigDev = CreateWithOneDockerConfig{
	Container: CreateWithOneDockerContainer{
		Image: utils.GetDefaultHelloImage(),
		Port:  80,
	},
	RequiredCpu:    "50m",
	RequiredMemory: "256Mi",
	RequiredDisk:   "50Mi",
	RequiredGpu:    false,
	NeedPg:         false,
	NeedRedis:      false,
	SshEnable:      false,
}

func CreateAppWithDevConfig(cfg *CreateDevContainerConfig, owner, name string) error {
	appPath := utils.GetAppPath(owner, name)
	localConfig := createConfigDev

	localConfig.Name = name
	localConfig.Title = cfg.Title
	if cfg != nil {
		if cfg.RequiredCpu != "" {
			localConfig.RequiredCpu = cfg.RequiredCpu
		}
		if cfg.RequiredMemory != "" {
			localConfig.RequiredMemory = cfg.RequiredMemory
		}
		if cfg.RequiredDisk != "" {
			localConfig.RequiredDisk = cfg.RequiredDisk
		}
		if cfg.RequiredGpu {
			localConfig.RequiredGpu = true
		}
		if cfg.GpuVendor != "" {
			localConfig.GpuVendor = cfg.GpuVendor
		}
		if len(cfg.ExposePorts) > 0 {
			localConfig.ExposePorts = cfg.ExposePorts
		}
		localConfig.SshEnable = cfg.SshEnable
	}
	at := AppTemplate{}
	at.WithDockerCfg(&localConfig).WithDockerDeployment(&localConfig).
		WithDockerService(&localConfig).WithDockerChartMetadata(&localConfig).WithDockerOwner(&localConfig)
	err := at.WriteDockerFile(&localConfig, appPath)
	if err != nil {
		return err
	}
	return nil
}

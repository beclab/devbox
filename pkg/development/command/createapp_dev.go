package command

type CreateDevContainerConfig struct {
	DevEnv         string `json:"devEnv"`
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

func CreateAppWithDevConfig(baseDir string, name string, cfg *CreateDevContainerConfig) error {
	createConfigDev.Name = name
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
	err := at.WriteDockerFile(createConfigDev, baseDir)
	if err != nil {
		return err
	}
	return nil
}

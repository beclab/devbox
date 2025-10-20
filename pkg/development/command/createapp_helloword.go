package command

import (
	"fmt"

	"github.com/beclab/devbox/pkg/utils"
	"github.com/containerd/containerd/reference/docker"
	"k8s.io/apimachinery/pkg/api/resource"
)

func CreateAppWithHelloWorldConfig(owner, name string, cfg *CreateWithHelloConfig) error {
	at := &AppTemplate{}
	createConfigExample := &CreateWithOneDockerConfig{
		Title:          cfg.Title,
		Name:           name,
		Container:      cfg.Container,
		RequiredCpu:    cfg.RequiredCpu,
		RequiredMemory: cfg.RequiredMemory,
	}

	if createConfigExample.Container.Image == "" {
		createConfigExample.Container.Image = utils.GetDefaultHelloImage()
	}
	if createConfigExample.Container.Port == 0 {
		createConfigExample.Container.Port = 80
	}
	if createConfigExample.RequiredCpu == "" {
		createConfigExample.RequiredCpu = "100m"
	}
	if createConfigExample.RequiredMemory == "" {
		createConfigExample.RequiredMemory = "128Mi"
	}
	_, err := docker.ParseDockerRef(createConfigExample.Container.Image)
	if err != nil {
		return fmt.Errorf("invalid image %v", err)
	}
	_, err = resource.ParseQuantity(createConfigExample.RequiredCpu)
	if err != nil {
		return fmt.Errorf("invalid requiredCpu %v", err)
	}
	_, err = resource.ParseQuantity(createConfigExample.RequiredMemory)
	if err != nil {
		return fmt.Errorf("invalid requiredMemory %v", err)
	}

	at.WithDockerCfg(createConfigExample).WithDockerDeployment(createConfigExample).
		WithDockerService(createConfigExample).WithDockerChartMetadata(createConfigExample).WithDockerOwner(createConfigExample)

	err = at.WriteDockerFile(createConfigExample, utils.GetAppPath(owner, name))
	if err != nil {
		return err
	}
	return nil
}

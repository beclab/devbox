package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	oac "github.com/beclab/Olares/framework/oac"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/manifest"
	"github.com/beclab/devbox/pkg/appcfg"
	"github.com/beclab/devbox/pkg/constants"
	"github.com/beclab/devbox/pkg/utils"

	jvalidator "github.com/go-playground/validator/v10"
	"helm.sh/helm/v3/pkg/chart"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//var vendorGpuMap = map[string]string{
//	"nvidia": "nvidia.com/gpu",
//	"amd":    "amd.com/gpu",
//	"intel":  "gpu.intel.com/i915",
//}

var validate = jvalidator.New()

type CreateWithOneDockerConfig struct {
	ID             string                       `json:"id"`
	Title          string                       `json:"title"`
	Name           string                       `json:"name" validate:"required,name"`
	Container      CreateWithOneDockerContainer `json:"container"`
	RequiredCpu    string                       `json:"requiredCpu" validate:"required,requiredCpu"`
	LimitedCpu     string                       `json:"limitedCpu" validate:"limitedCpu"`
	RequiredMemory string                       `json:"requiredMemory" validate:"required,requiredMemory"`
	LimitedMemory  string                       `json:"limitedMemory" validate:"limitedMemory"`
	RequiredDisk   string                       `json:"requiredDisk" validate:"requiredDisk"`
	LimitedDisk    string                       `json:"limitedDisk" validate:"limitedDisk"`
	RequiredGpu    bool                         `json:"requiredGpu"`
	GpuVendor      string                       `json:"gpuVendor" validate:"gpuVendor"`
	NeedPg         bool                         `json:"needPg"`
	NeedRedis      bool                         `json:"needRedis"`
	Env            map[string]string            `json:"env"`
	Mounts         map[string]string            `json:"mounts"`
	ExposePorts    string                       `json:"exposePorts"`
	SshEnable      bool                         `json:"sshEnable"`
}

type CreateWithOneDockerContainer struct {
	Image        string `json:"image" validate:"required,image"`
	StartCmd     string `json:"startCmd"`
	StartCmdArgs string `json:"startCmdArgs"`
	Port         int    `json:"port"`
}

type CreateWithHelloConfig struct {
	Title          string                       `json:"title"`
	Container      CreateWithOneDockerContainer `json:"container"`
	RequiredCpu    string                       `json:"requiredCpu"`
	RequiredMemory string                       `json:"requiredMemory"`
}

type createWithOneDocker struct {
	baseCommand
}

func CreateWithOneDocker() *createWithOneDocker {
	return &createWithOneDocker{
		*newBaseCommand(),
	}
}

func (c *createWithOneDocker) WithDir(dir string) *createWithOneDocker {
	c.baseCommand.withDir(dir)
	return c
}

func (c *createWithOneDocker) Run(cfg *CreateWithOneDockerConfig, owner string) error {
	at := NewAppTemplate(context.Background())
	at.WithDockerCfg(cfg).WithDockerDeployment(cfg).WithDockerService(cfg).WithDockerChartMetadata(cfg).WithDockerOwner(cfg)

	appPath := utils.GetAppPath(owner, cfg.Name)

	return at.WriteDockerFile(cfg, appPath)
}

func (at *AppTemplate) checkMountPath(mounts map[string]string, prefix string) bool {

	for key := range mounts {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	return false
}

func (at *AppTemplate) WithDockerCfg(config *CreateWithOneDockerConfig) *AppTemplate {
	configType := "app"

	appcfg := oac.AppConfiguration{
		APIVersion:    "v1",
		ConfigVersion: configVersionFor(at.useNewSchema),
		ConfigType:    configType,
		Metadata: manifest.AppMetaData{
			Name:        config.Name,
			Icon:        defaultIcon,
			Description: fmt.Sprintf("app %s", config.Name),
			AppID:       config.Name,
			Version:     "0.0.1",
			Title:       config.Title,
			Categories:  []string{"Utilities"},
		},
		Spec: manifest.AppSpec{
			VersionName: "0.0.1",
			SupportArch: []string{"amd64", "arm64"},
		},
	}

	appcfg.Permission.AppData = at.checkMountPath(config.Mounts, "/app/data/")
	appcfg.Permission.AppCache = at.checkMountPath(config.Mounts, "/app/cache/")
	//  {{ .Values.sharedlib }}
	appcfg.Permission.UserData = make([]string, 0)
	if at.checkMountPath(config.Mounts, "/Home/") {
		for key := range config.Mounts {
			if strings.HasPrefix(key, "/Home/") {
				appcfg.Permission.UserData = append(appcfg.Permission.UserData, key)
			}
		}
	}

	//middleware := manifest.Middleware{}
	appcfg.Middleware = &manifest.Middleware{}
	if config.NeedRedis {
		appcfg.Middleware.Redis = &manifest.RedisConfig{
			Namespace: "redis",
		}
	}

	if config.NeedPg {
		appcfg.Middleware.Postgres = &manifest.PostgresConfig{
			Username: "postgres",
			Databases: []manifest.Database{
				{
					Name:        config.Name,
					Distributed: true,
				},
			},
		}
	}

	entrances := make([]appv1.Entrance, 0)
	name := config.Name
	//port := config.Container.Port

	entrances = append(entrances, appv1.Entrance{
		Name:       name,
		Host:       name,
		Port:       int32(config.Container.Port),
		Title:      config.Name,
		Icon:       defaultIcon,
		AuthLevel:  "private",
		OpenMethod: "default",
	})
	exposePorts := strings.Split(config.ExposePorts, ",")
	for _, portStr := range exposePorts {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}
		entrances = append(entrances, appv1.Entrance{
			Name:      fmt.Sprintf("%s-dev-%s", name, portStr),
			Host:      name,
			Port:      int32(port),
			Title:     fmt.Sprintf("%s-dev-%s", name, portStr),
			Icon:      defaultIcon,
			AuthLevel: "private",
			Skip:      true,
			Invisible: true,
		})
	}

	appcfg.Entrances = entrances

	if config.SshEnable {
		appcfg.Ports = append(appcfg.Ports, appv1.ServicePort{
			Name:              "ssh",
			Host:              name,
			Port:              22,
			AddToTailscaleAcl: true,
		})
	}

	res := oac.ManifestResourceLimits{
		RequiredCPU:    config.RequiredCpu,
		RequiredMemory: config.RequiredMemory,
		RequiredDisk:   "50Mi",
		LimitedDisk:    "5Gi",
		LimitedCPU:     config.LimitedCpu,
		LimitedMemory:  config.LimitedMemory,
	}
	if res.LimitedCPU == "" {
		res.LimitedCPU = res.RequiredCPU
	}
	if res.LimitedMemory == "" {
		res.LimitedMemory = res.RequiredMemory
	}
	if config.RequiredDisk != "" {
		res.RequiredDisk = config.RequiredDisk
	}

	requiredCPU, _ := resource.ParseQuantity(res.RequiredCPU)
	requiredMemory, _ := resource.ParseQuantity(res.RequiredMemory)
	limitedCPU, _ := resource.ParseQuantity(res.LimitedCPU)
	limitedMemory, _ := resource.ParseQuantity(res.LimitedMemory)

	if requiredCPU.Cmp(limitedCPU) > 0 {
		res.LimitedCPU = res.RequiredCPU
	}
	if requiredMemory.Cmp(limitedMemory) > 0 {
		res.LimitedMemory = res.RequiredMemory
	}

	if config.RequiredGpu {
		res.RequiredGPU = "1Gi"
		gpuType, _ := utils.GetFirstClusterGpuType(context.Background())
		applyAppGpuMode(&appcfg.Spec, at.useNewSchema, gpuType, res)
	} else {
		applyAppResources(&appcfg.Spec, at.useNewSchema, res)
	}
	appcfg.Options.Dependencies = []manifest.Dependency{
		{
			Name:    "olares",
			Type:    "system",
			Version: constants.SupportOsVersion,
		},
	}
	at.appCfg = &appcfg
	return at
}

func (at *AppTemplate) WithDockerDeployment(config *CreateWithOneDockerConfig) *AppTemplate {
	replicas := int32(1)
	gpuType, _ := utils.GetFirstClusterGpuType(context.Background())
	res := extractAppResources(at.appCfg.Spec, gpuType)
	requestCPU, _ := resource.ParseQuantity(res.RequiredCPU)
	requestMemory, _ := resource.ParseQuantity(res.RequiredMemory)
	limitedCPU, _ := resource.ParseQuantity(res.LimitedCPU)
	limitedMemory, _ := resource.ParseQuantity(res.LimitedMemory)

	klog.Infof("WithDockerDeployment: requestCPU: %v", requestCPU.String())
	klog.Infof("WithDockerDeployment: limitedCPU: %v", limitedCPU.String())

	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: "{{ .Release.Namespace }}",
			Labels: map[string]string{
				"io.kompose.service": config.Name,
			},
			Annotations: func() map[string]string {
				anno := map[string]string{}
				if gpuType != "" && config.RequiredGpu {
					anno[constants.ApplicationGpuInjectKey] = "true"
				}
				return anno
			}(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"io.kompose.service": config.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"io.kompose.network/chrome-default": "true",
						"io.kompose.service":                config.Name,
					},
					Annotations: map[string]string{
						constants.ExposePortsLabel: config.ExposePorts,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  config.Name,
							Image: config.Container.Image,
							//Command: []string{config.Container.StartCmd},
							//Args:    []string{config.Container.StartCmdArgs},
							Resources: corev1.ResourceRequirements{
								Requests: map[corev1.ResourceName]resource.Quantity{
									corev1.ResourceCPU:    requestCPU,
									corev1.ResourceMemory: requestMemory,
								},
								Limits: map[corev1.ResourceName]resource.Quantity{
									corev1.ResourceCPU:    limitedCPU,
									corev1.ResourceMemory: limitedMemory,
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}
	if len(config.Container.StartCmd) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Command = ParseCommand(config.Container.StartCmd)
	}
	if len(config.Container.StartCmdArgs) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Args = []string{config.Container.StartCmdArgs}
	}

	if config.ExposePorts != "" {
		exposePorts := strings.Split(config.ExposePorts, ",")
		allDomainConfigs := make([]appcfg.DefaultThirdLevelDomainConfig, 0)
		for _, portStr := range exposePorts {
			_, err := strconv.Atoi(portStr)
			if err != nil {
				continue
			}
			allDomainConfigs = append(allDomainConfigs, appcfg.DefaultThirdLevelDomainConfig{
				AppName:      config.Name,
				EntranceName: fmt.Sprintf("%s-dev-%s", config.Name, portStr),
				//ThirdLevelDomain: fmt.Sprintf("%s-%s", utils.GetAppID(config.Name), portStr),
			})
		}
		if len(allDomainConfigs) > 0 {
			allDomainConfigStr, _ := json.Marshal(allDomainConfigs)
			deployment.Annotations[constants.ApplicationDefaultThirdLevelDomain] = string(allDomainConfigStr)
		}
	}

	ports := make([]corev1.ContainerPort, 0)
	ports = append(ports, corev1.ContainerPort{
		ContainerPort: int32(config.Container.Port),
	})

	deployment.Spec.Template.Spec.Containers[0].Ports = ports

	env := []corev1.EnvVar{
		{
			Name:  "PGID",
			Value: "1000",
		},
		{
			Name:  "PUID",
			Value: "1000",
		},
		{
			Name:  "TZ",
			Value: "Etc/UTC",
		},
	}
	envMap := make(map[string]int)

	if config.NeedPg {
		postgresEnv := []corev1.EnvVar{
			{
				Name:  "PG_HOST",
				Value: "{{ .Values.postgres.host }}",
			},
			{
				Name:  "PG_PORT",
				Value: "{{ .Values.postgres.port }}",
			},
			{
				Name:  "PG_USER",
				Value: "{{ .Values.postgres.username }}",
			},
			{
				Name:  "PG_PASS",
				Value: "{{ .Values.postgres.password }}",
			},
			{
				Name:  "PG_DBNAME",
				Value: "{{ .Values.postgres.databases." + config.Name + " }}",
			},
		}
		env = append(env, postgresEnv...)
	}
	if config.NeedRedis {
		redisEnv := []corev1.EnvVar{
			{
				Name:  "REDIS_HOST",
				Value: "{{ .Values.redis.host }}",
			},
			{
				Name:  "REDIS_PORT",
				Value: "{{ .Values.redis.port }}",
			},
			{
				Name:  "REDIS_USER",
				Value: "{{ .Values.redis.username }}",
			},
			{
				Name:  "REDIS_PASS",
				Value: "{{ .Values.redis.password }}",
			},
		}
		env = append(env, redisEnv...)
	}
	for i, e := range env {
		envMap[e.Name] = i
	}
	if config.Env != nil {
		for name, value := range config.Env {
			if idx, exists := envMap[name]; exists {
				env[idx].Value = value
			} else {
				env = append(env, corev1.EnvVar{Name: name, Value: value})
			}
		}
	}
	deployment.Spec.Template.Spec.Containers[0].Env = env

	volumeMounts := make([]corev1.VolumeMount, 0)

	volumes := make([]corev1.Volume, 0)
	t := corev1.HostPathDirectoryOrCreate

	for hostPath, mountPath := range config.Mounts {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      formatPathToVolumeName(hostPath),
			MountPath: mountPath,
		})
		volumes = append(volumes, corev1.Volume{
			Name: formatPathToVolumeName(hostPath),
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Type: &t,
					Path: replacePath(hostPath, config.Name),
				},
			},
		})
	}
	if len(volumeMounts) > 0 {
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = volumeMounts
	}

	if len(volumes) > 0 {
		deployment.Spec.Template.Spec.Volumes = volumes
	}
	at.deployment = &deployment
	return at
}

func (at *AppTemplate) WithDockerService(config *CreateWithOneDockerConfig) *AppTemplate {
	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"io.kompose.service": config.Name,
			},
			Name:      config.Name,
			Namespace: "{{ .Release.Namespace }}",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"io.kompose.service": config.Name,
			},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{},
		},
	}
	ports := make([]corev1.ServicePort, 0)
	ports = append(ports, corev1.ServicePort{
		Name:       strconv.Itoa(config.Container.Port),
		Port:       int32(config.Container.Port),
		TargetPort: intstr.Parse(strconv.Itoa(config.Container.Port)),
	})

	for _, portStr := range strings.Split(config.ExposePorts, ",") {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}
		ports = append(ports, corev1.ServicePort{
			Name:       fmt.Sprintf("%s-dev-%s", config.Name, portStr),
			Port:       int32(port),
			TargetPort: intstr.Parse(portStr),
		})
	}
	if config.SshEnable {
		ports = append(ports, corev1.ServicePort{
			Name:       "ssh",
			Port:       22,
			TargetPort: intstr.Parse("22"),
		})
	}

	if len(ports) > 0 {
		service.Spec.Ports = ports
	}
	at.services = append(at.services, &service)
	return at
}

func (at *AppTemplate) WithDockerChartMetadata(cfg *CreateWithOneDockerConfig) *AppTemplate {
	metadata := chart.Metadata{
		APIVersion:  "v2",
		Name:        cfg.Name,
		Description: "description",
		Type:        "application",
		Version:     "0.0.1",
		AppVersion:  "0.0.1",
	}
	at.chartMetadata = &metadata
	return at
}

func (at *AppTemplate) WithDockerOwner(cfg *CreateWithOneDockerConfig) *AppTemplate {
	at.owner = &Owner{}
	return at
}

func (at *AppTemplate) WriteDockerFile(cfg *CreateWithOneDockerConfig, path string) (err error) {
	if existDir(path) {
		return os.ErrExist
	}
	createPath := filepath.Join(path, "templates")
	err = os.MkdirAll(createPath, os.ModePerm)
	if err != nil {
		return err
	}
	if at.appCfg != nil {
		yml, err := ToYaml(at.appCfg)

		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join(path, constants.AppCfgFileName), yml, 0644)
		if err != nil {
			return err
		}
	}
	if at.chartMetadata != nil {
		yml, err := ToYaml(at.chartMetadata)

		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join(path, "Chart.yaml"), yml, 0644)
		if err != nil {
			return err
		}
	}

	if at.owner != nil {
		err = ioutil.WriteFile(filepath.Join(path, "owners"), []byte{}, 0644)
		if err != nil {
			return err
		}
	}
	var yml []byte
	if at.deployment != nil {
		yml, err = ToYaml(at.deployment)
		if err != nil {
			return err
		}

	}
	var sep = []byte("\n---\n")
	for _, svc := range at.services {
		serviceYml, err := ToYaml(svc)
		if err != nil {
			return err
		}
		yml = append(yml, sep...)
		yml = append(yml, serviceYml...)
	}
	err = ioutil.WriteFile(filepath.Join(path, "templates", "deployment.yaml"), yml, 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(path, "values.yaml"), nil, 0644)
	if err != nil {
		return err
	}

	return err
}

func ParseCommand(cmd string) []string {
	if cmd == "" {
		return []string{}
	}

	var result []string
	var current strings.Builder
	var inQuotes bool
	var quoteChar rune

	cmd = strings.TrimSpace(cmd)

	for _, char := range cmd {
		switch {
		case char == '"' || char == '\'':
			if inQuotes && char == quoteChar {
				inQuotes = false
			} else if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else {
				current.WriteRune(char)
			}
		case unicode.IsSpace(char):
			if inQuotes {
				current.WriteRune(char)
			} else if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(char)
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

func formatPathToVolumeName(path string) string {
	trimmed := strings.Trim(path, "/")
	result := strings.ToLower(strings.ReplaceAll(trimmed, "/", "-"))
	return result

}

func replacePath(input string, name string) string {
	replacements := map[string]string{
		"/app/data/":  "{{ .Values.userspace.appData }}/" + name + "/",
		"/app/cache/": "{{ .Values.userspace.appCache }}/" + name + "/",
		"/Home/":      "{{ .Values.userspace.userData }}/",
	}

	result := input
	for oldPath, newPath := range replacements {
		result = strings.Replace(result, oldPath, newPath, -1)
	}

	return result
}

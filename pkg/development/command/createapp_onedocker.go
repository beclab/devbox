package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/beclab/devbox/pkg/constants"
	"github.com/beclab/oachecker"

	jvalidator "github.com/go-playground/validator/v10"
	"helm.sh/helm/v3/pkg/chart"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var validate = jvalidator.New()

type CreateWithOneDockerConfig struct {
	ID             string                       `json:"id"`
	Name           string                       `json:"name" validate:"required,name"`
	Container      CreateWithOneDockerContainer `json:"container"`
	RequiredCpu    string                       `json:"requiredCpu" validate:"required,requiredCpu"`
	LimitedCpu     string                       `json:"limitedCpu" validate:"limitedCpu"`
	RequiredMemory string                       `json:"requiredMemory" validate:"required,requiredMemory"`
	LimitedMemory  string                       `json:"limitedMemory" validate:"limitedMemory"`
	RequiredGpu    bool                         `json:"requiredGpu"`
	NeedPg         bool                         `json:"needPg"`
	NeedRedis      bool                         `json:"needRedis"`
	Env            map[string]string            `json:"env"`
	Mounts         map[string]string            `json:"mounts"`
}

type CreateWithOneDockerContainer struct {
	Image        string `json:"image" validate:"required,image"`
	StartCmd     string `json:"startCmd"`
	StartCmdArgs string `json:"startCmdArgs"`
	Port         int    `json:"port"`
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

func (c *createWithOneDocker) Run(cfg *CreateWithOneDockerConfig) error {
	at := AppTemplate{}
	at.WithDockerCfg(cfg).WithDockerDeployment(cfg).WithDockerDeployment(cfg).WithDockerService(cfg).WithDockerChartMetadata(cfg).WithDockerOwner(cfg)

	baseDir := c.dir
	if baseDir == "" {
		baseDir = os.Getenv("BASE_DIR")
		if baseDir == "" {
			baseDir = "/tmp"
		}
	}

	return at.WriteDockerFile(cfg, baseDir)
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

	appcfg := oachecker.AppConfiguration{
		ConfigVersion: "v1",
		ConfigType:    configType,
		Metadata: oachecker.AppMetaData{
			Name:        config.Name,
			Icon:        defaultIcon,
			Description: fmt.Sprintf("app %s", config.Name),
			AppID:       config.Name,
			Version:     "0.0.1",
			Title:       config.Name,
			Categories:  []string{"dev"},
		},
		Spec: oachecker.AppSpec{
			RequiredMemory: config.RequiredMemory,
			RequiredCPU:    config.RequiredCpu,
			RequiredDisk:   "50Mi",
			LimitedMemory:  config.LimitedMemory,
			LimitedCPU:     config.LimitedCpu,
			VersionName:    "0.0.1",
			SupportArch:    []string{"amd64"},
		},
	}

	appcfg.Permission.AppData = at.checkMountPath(config.Mounts, "{{ .Values.userspace.appData }}")
	appcfg.Permission.AppCache = at.checkMountPath(config.Mounts, "{{ .Values.userspace.appCache }}")
	//  {{ .Values.sharedlib }}
	appcfg.Permission.UserData = make([]string, 0)
	if at.checkMountPath(config.Mounts, "{{ .Values.userspace.userData }}") {
		for key := range config.Mounts {
			if strings.HasPrefix(key, "{{ .Values.userspace.userData }})") {
				appcfg.Permission.UserData = append(appcfg.Permission.UserData, config.Mounts[key])
			}
		}
	}

	//middleware := oachecker.Middleware{}
	appcfg.Middleware = &oachecker.Middleware{}
	if config.NeedRedis {
		appcfg.Middleware.Redis = &oachecker.RedisConfig{
			Namespace: "redis",
		}
	}

	if config.NeedPg {
		appcfg.Middleware.Postgres = &oachecker.PostgresConfig{
			Username: "postgres",
			Databases: []oachecker.Database{
				{
					Name:        config.Name,
					Distributed: true,
				},
			},
		}
	}

	entrances := make([]oachecker.Entrance, 0)
	name := config.Name
	port := config.Container.Port

	entrances = append(entrances, oachecker.Entrance{
		Name:       name,
		Host:       name,
		Port:       int32(port),
		Title:      config.Name,
		Icon:       defaultIcon,
		AuthLevel:  "private",
		OpenMethod: "default",
	})

	appcfg.Entrances = entrances

	if config.RequiredGpu {
		appcfg.Spec.RequiredGPU = "1"
	} else {
		appcfg.Spec.RequiredGPU = "0"
	}

	appcfg.Spec.RequiredCPU = config.RequiredCpu
	appcfg.Spec.RequiredMemory = config.RequiredMemory
	if config.LimitedCpu == "" {
		appcfg.Spec.LimitedCPU = config.RequiredCpu
	} else {
		appcfg.Spec.LimitedCPU = config.LimitedCpu

	}
	if config.LimitedMemory == "" {
		appcfg.Spec.LimitedMemory = config.RequiredMemory
	} else {
		appcfg.Spec.LimitedMemory = config.LimitedMemory
	}

	requiredCPU, _ := resource.ParseQuantity(appcfg.Spec.RequiredCPU)
	requiredMemory, _ := resource.ParseQuantity(appcfg.Spec.RequiredMemory)
	limitedCPU, _ := resource.ParseQuantity(appcfg.Spec.LimitedCPU)
	limitedMemory, _ := resource.ParseQuantity(appcfg.Spec.LimitedMemory)

	if requiredCPU.Cmp(limitedCPU) > 0 {
		appcfg.Spec.LimitedCPU = appcfg.Spec.RequiredCPU
	}

	if requiredMemory.Cmp(limitedMemory) > 0 {
		appcfg.Spec.LimitedMemory = appcfg.Spec.RequiredMemory
	}

	//if cfg.OSVersion != "" {
	//	appcfg.Options.Dependencies = []application.Dependency{
	//		{
	//			Name:    "terminus",
	//			Type:    "system",
	//			Version: cfg.OSVersion,
	//		},
	//	}
	//}
	at.appCfg = &appcfg
	return at
}

func (at *AppTemplate) WithDockerDeployment(config *CreateWithOneDockerConfig) *AppTemplate {
	replicas := int32(1)
	requestCPU, _ := resource.ParseQuantity(at.appCfg.Spec.RequiredCPU)
	requestMemory, _ := resource.ParseQuantity(at.appCfg.Spec.RequiredMemory)
	limitedCPU, _ := resource.ParseQuantity(at.appCfg.Spec.LimitedCPU)
	limitedMemory, _ := resource.ParseQuantity(at.appCfg.Spec.LimitedMemory)

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

	ports := make([]corev1.ContainerPort, 0)
	ports = append(ports, corev1.ContainerPort{
		ContainerPort: int32(config.Container.Port),
	})
	//}
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
	if config.Env != nil {
		for name, value := range config.Env {
			env = append(env, corev1.EnvVar{Name: name, Value: value})
		}
	}
	deployment.Spec.Template.Spec.Containers[0].Env = env

	volumeMounts := make([]corev1.VolumeMount, 0)
	//klog.Info("cfg.AppCache: ", cfg.AppCache)
	if at.appCfg.Permission.AppCache {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "appcache",
			MountPath: "/appcache",
		})
	}
	if at.appCfg.Permission.AppData {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "appdata",
			MountPath: "/appdata",
		})
	}
	if len(at.appCfg.Permission.UserData) > 0 {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "userdata",
			MountPath: "/userdata",
		})
	}
	if len(volumeMounts) > 0 {
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = volumeMounts
	}

	volumes := make([]corev1.Volume, 0)
	t := corev1.HostPathDirectoryOrCreate
	if at.appCfg.Permission.AppCache {
		volumes = append(volumes, corev1.Volume{
			Name: "appcache",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Type: &t,
					Path: "{{ .Values.userspace.appCache }}/" + config.Name,
				},
			},
		})
	}

	if at.appCfg.Permission.AppData {
		volumes = append(volumes, corev1.Volume{
			Name: "appdata",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Type: &t,
					Path: "{{ .Values.userspace.appData }}/" + config.Name,
				},
			},
		})
	}

	for key := range at.appCfg.Permission.UserData {
		volumes = append(volumes, corev1.Volume{
			Name: "userdata",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Type: &t,
					Path: "{{ .Values.userspace.userData }}" + "/" + at.appCfg.Permission.UserData[key],
				},
			},
		})
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
	if len(ports) > 0 {
		service.Spec.Ports = ports
	}
	at.service = &service
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

func (at *AppTemplate) WriteDockerFile(cfg *CreateWithOneDockerConfig, baseDir string) (err error) {
	path := filepath.Join(baseDir, cfg.Name)
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
	if at.service != nil {
		serviceYml, err := ToYaml(at.service)
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

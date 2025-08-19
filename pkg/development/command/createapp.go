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

	"github.com/beclab/devbox/pkg/utils"
	"github.com/beclab/oachecker"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/devbox/pkg/constants"
	"github.com/jedib0t/go-pretty/v6/table"
	"helm.sh/helm/v3/pkg/chart"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

// const defaultIcon = "https://file.bttcdn.com/appstore/accounts/icon.png"
const (
	defaultIcon           = "https://file.bttcdn.com/appstore/default/defaulticon.webp"
	traefik               = "traefik"
	appKubernetesName     = "app.kubernetes.io/name"
	appKubernetesInstance = "app.kubernetes.io/instance"
)

type CreateConfig struct {
	Name string `json:"name"`
	Type string `json:"type"`

	SystemDB   bool `json:"systemDB,omitempty"`
	Redis      bool `json:"redis,omitempty"`
	MongoDB    bool `json:"mongodb,omitempty"`
	PostgreSQL bool `json:"postgreSQL,omitempty"`

	SystemCall bool `json:"systemCall,omitempty"`

	Img   string `json:"img"`
	Ports []int  `json:"ports,omitempty"`

	IngressRouter bool   `json:"ingressRouter,omitempty"`
	Traefik       bool   `json:"traefik,omitempty"`
	WebsitePort   string `json:"websitePort,omitempty"`

	AppData  bool     `json:"appData,omitempty"`
	AppCache bool     `json:"appCache,omitempty"`
	UserData []string `json:"userData,omitempty"`

	NeedGPU        bool   `json:"needGpu,omitempty"`
	RequiredGPU    string `json:"requiredGpu,omitempty"`
	RequiredMemory string `json:"requiredMemory,omitempty"`
	RequiredDisk   string `json:"requiredDisk,omitempty"`
	RequiredCPU    string `json:"requiredCpu,omitempty"`
	LimitedMemory  string `json:"limitedMemory,omitempty"`
	LimitedCPU     string `json:"limitedCpu,omitempty"`
	OSVersion      string `json:"osVersion"`

	DevEnv string `json:"devEnv"`
}

type createApp struct {
	baseCommand
}

func CreateApp() *createApp {
	return &createApp{
		*newBaseCommand(),
	}
}

func (c *createApp) WithDir(dir string) *createApp {
	c.baseCommand.withDir(dir)
	return c
}

func (c *createApp) Run(ctx context.Context, cfg *CreateConfig, owner string) error {
	at := AppTemplate{}
	at.WithAppCfg(cfg).WithDeployment(cfg).WithService(cfg).WithChartMetadata(cfg).WithOwner(cfg)
	if cfg.Traefik {
		at.WithTraefik(cfg)
	}

	return at.WriteFile(cfg, owner)
}

type AppTemplate struct {
	appCfg        *oachecker.AppConfiguration
	deployment    *appsv1.Deployment
	service       *corev1.Service
	chartMetadata *chart.Metadata
	traefik       *Traefik
	owner         *Owner
}

type Owner struct {
}

type Traefik struct {
	sa          *corev1.ServiceAccount
	pvc         *corev1.PersistentVolumeClaim
	role        *rbacv1.Role
	roleBinding *rbacv1.RoleBinding
	svc         *corev1.Service
	deployment  *appsv1.Deployment
}

func (at *AppTemplate) WithAppCfg(cfg *CreateConfig) *AppTemplate {
	appRef := make([]string, 0)
	configType := cfg.Type
	if configType == "" {
		configType = "app"
	}
	appcfg := oachecker.AppConfiguration{
		ConfigVersion: "0.8.0",
		ConfigType:    configType,
		Metadata: oachecker.AppMetaData{
			Name:        cfg.Name,
			Icon:        defaultIcon,
			Description: fmt.Sprintf("app %s", cfg.Name),
			AppID:       cfg.Name,
			Version:     "0.0.1",
			Title:       cfg.Name,
			Categories:  []string{"dev"},
		},
		Spec: oachecker.AppSpec{
			RequiredMemory: "100Mi",
			RequiredCPU:    "50m",
			RequiredDisk:   "50Mi",
			LimitedMemory:  "1000Mi",
			LimitedCPU:     "1000m",
			VersionName:    "0.0.1",
			SupportArch:    []string{"amd64"},
		},
		Options: oachecker.Options{
			AppScope: &oachecker.AppScope{
				AppRef: appRef,
			},
		},
	}
	entrances := make([]oachecker.Entrance, 0)
	name := cfg.Name
	port, _ := strconv.Atoi(cfg.WebsitePort)

	if cfg.IngressRouter {
		if cfg.Traefik {
			name = "traefik"
		} else {
			name = ""
		}
	}
	entrances = append(entrances, oachecker.Entrance{
		Name:       name,
		Host:       name,
		Port:       int32(port),
		Title:      cfg.Name,
		Icon:       defaultIcon,
		AuthLevel:  "private",
		OpenMethod: "default",
	})

	appcfg.Entrances = entrances

	if cfg.AppData {
		appcfg.Permission.AppData = true
	}
	if cfg.AppCache {
		appcfg.Permission.AppCache = true
	}
	if len(cfg.UserData) > 0 {
		appcfg.Permission.UserData = cfg.UserData
	} else {
		appcfg.Permission.UserData = make([]string, 0)
	}

	middleware := oachecker.Middleware{}
	if cfg.SystemDB {
		if cfg.Redis {
			middleware.Redis = &oachecker.RedisConfig{
				Namespace: "redis",
			}
		}
		if cfg.MongoDB {
			middleware.MongoDB = &oachecker.MongodbConfig{
				Username: "root",
				Databases: []oachecker.Database{
					{
						Name: cfg.Name,
					},
				},
			}
		}
		if cfg.PostgreSQL {
			middleware.Postgres = &oachecker.PostgresConfig{
				Username: "postgres",
				Databases: []oachecker.Database{
					{
						Name:        cfg.Name,
						Distributed: true,
					},
				},
			}
		}
	}

	if cfg.NeedGPU && cfg.RequiredGPU != "" {
		appcfg.Spec.RequiredGPU = cfg.RequiredGPU
	}
	if cfg.RequiredMemory != "" {
		appcfg.Spec.RequiredMemory = cfg.RequiredMemory
	}

	if cfg.RequiredDisk != "" {
		appcfg.Spec.RequiredDisk = cfg.RequiredDisk
	}

	if cfg.RequiredCPU != "" {
		appcfg.Spec.RequiredCPU = cfg.RequiredCPU
	}

	if cfg.LimitedMemory != "" {
		appcfg.Spec.LimitedMemory = cfg.LimitedMemory
	}

	if cfg.LimitedCPU != "" {
		appcfg.Spec.LimitedCPU = cfg.LimitedCPU
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

	if cfg.OSVersion != "" {
		if appcfg.Options.Dependencies == nil {
			dependencies := make([]oachecker.Dependency, 0)
			appcfg.Options.Dependencies = &dependencies
		}
		*appcfg.Options.Dependencies = []oachecker.Dependency{
			{
				Name:    "terminus",
				Type:    "system",
				Version: cfg.OSVersion,
			},
		}
	}
	at.appCfg = &appcfg
	return at
}

func (at *AppTemplate) WithDeployment(cfg *CreateConfig) *AppTemplate {
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
			Name:      cfg.Name,
			Namespace: "{{ .Release.Namespace }}",
			Labels: map[string]string{
				"io.kompose.service": cfg.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"io.kompose.service": cfg.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"io.kompose.network/chrome-default": "true",
						"io.kompose.service":                cfg.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  cfg.Name,
							Image: cfg.Img,
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
	ports := make([]corev1.ContainerPort, 0)
	for _, port := range cfg.Ports {
		ports = append(ports, corev1.ContainerPort{
			ContainerPort: int32(port),
		})
	}
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
	if cfg.MongoDB {
		mongoEnv := []corev1.EnvVar{
			{
				Name:  "MONGODB_HOST",
				Value: "{{ .Values.mongodb.host }}",
			},
			{
				Name:  "MONGODB_PORT",
				Value: "{{ .Values.mongodb.port }}",
			},
			{
				Name:  "MONGODB_USER",
				Value: "{{ .Values.mongodb.username }}",
			},
			{
				Name:  "MONGODB_PASS",
				Value: "{{ .Values.mongodb.password }}",
			},
			{
				Name:  "MONGODB_DBNAME",
				Value: "{{ .Values.mongodb.databases." + cfg.Name + " }}",
			},
		}
		env = append(env, mongoEnv...)
	}
	if cfg.PostgreSQL {
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
				Value: "{{ .Values.postgres.databases." + cfg.Name + " }}",
			},
		}
		env = append(env, postgresEnv...)
	}
	if cfg.Redis {
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
	deployment.Spec.Template.Spec.Containers[0].Env = env

	volumeMounts := make([]corev1.VolumeMount, 0)
	klog.Info("cfg.AppCache: ", cfg.AppCache)
	//if cfg.AppCache {
	//	volumeMounts = append(volumeMounts, corev1.VolumeMount{
	//		Name:      "appcache",
	//		MountPath: "/appcache",
	//	})
	//}
	//if len(cfg.UserData) > 0 {
	//	volumeMounts = append(volumeMounts, corev1.VolumeMount{
	//		Name:      "userdata",
	//		MountPath: "/userdata",
	//	})
	//}
	if len(volumeMounts) > 0 {
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = volumeMounts
	}

	volumes := make([]corev1.Volume, 0)
	//t := corev1.HostPathDirectoryOrCreate
	//if cfg.AppCache {
	//	volumes = append(volumes, corev1.Volume{
	//		Name: "appcache",
	//		VolumeSource: corev1.VolumeSource{
	//			HostPath: &corev1.HostPathVolumeSource{
	//				Type: &t,
	//				Path: "{{ .Values.userspace.appCache }}/" + cfg.Name,
	//			},
	//		},
	//	})
	//}

	//if len(cfg.UserData) > 0 {
	//	volumes = append(volumes, corev1.Volume{
	//		Name: "userdata",
	//		VolumeSource: corev1.VolumeSource{
	//			HostPath: &corev1.HostPathVolumeSource{
	//				Type: &t,
	//				Path: "{{ .Values.userspace.userData }}",
	//			},
	//		},
	//	})
	//}
	if len(volumes) > 0 {
		deployment.Spec.Template.Spec.Volumes = volumes
	}
	at.deployment = &deployment
	return at

}

func (at *AppTemplate) WithService(cfg *CreateConfig) *AppTemplate {
	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"io.kompose.service": cfg.Name,
			},
			Name:      cfg.Name,
			Namespace: "{{ .Release.Namespace }}",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"io.kompose.service": cfg.Name,
			},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{},
		},
	}
	ports := make([]corev1.ServicePort, 0)
	for _, port := range cfg.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       strconv.Itoa(port),
			Port:       int32(port),
			TargetPort: intstr.Parse(strconv.Itoa(port)),
		})
	}
	if len(ports) > 0 {
		service.Spec.Ports = ports
	}
	at.service = &service
	return at
}

func (at *AppTemplate) WithChartMetadata(cfg *CreateConfig) *AppTemplate {
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

func (at *AppTemplate) WithOwner(cfg *CreateConfig) *AppTemplate {
	at.owner = &Owner{}
	return at
}

func (at *AppTemplate) WithTraefik(cfg *CreateConfig) *AppTemplate {
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      traefik,
			Namespace: "{{ .Release.Namespace }}",
			Labels: map[string]string{
				appKubernetesName:     traefik,
				appKubernetesInstance: traefik + "-" + "{{ .Release.Namespace }}",
			},
		},
	}
	storage, _ := resource.ParseQuantity("5G")
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      traefik,
			Namespace: "{{ .Release.Namespace }}",
			Labels: map[string]string{
				appKubernetesName:     traefik,
				appKubernetesInstance: traefik + "-" + "{{ .Release.Namespace }}",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: storage,
				},
			},
		},
	}
	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      traefik + "-" + "{{ .Release.Namespace }}",
			Namespace: "{{ .Release.Namespace }}",
			Labels: map[string]string{
				appKubernetesName:     traefik,
				appKubernetesInstance: traefik + "-" + "{{ .Release.Namespace }}",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"extensions",
					"networking.k8s.io",
				},
				Resources: []string{
					"ingressclasses",
					"ingresses",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services", "endpoints", "secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"extensions", "networking.k8s.io"},
				Resources: []string{"ingresses/status"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"traefik.containo.us"},
				Resources: []string{
					"ingressroutes",
					"ingressroutetcps",
					"ingressrouteudps",
					"middlewares",
					"middlewaretcps",
					"tlsoptions",
					"tlsstores",
					"traefikservices",
					"serverstransports",
				},
				Verbs: []string{"get", "list", "watch"},
			},
		},
	}

	roleBinding := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      traefik + "-" + "{{ .Release.Namespace }}",
			Namespace: "{{ .Release.Namespace }}",
			Labels: map[string]string{
				appKubernetesName:     traefik,
				appKubernetesInstance: traefik + "-" + "{{ .Release.Namespace }}",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.i",
			Kind:     "Role",
			Name:     traefik + "-" + "{{ .Release.Namespace }}",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      traefik,
				Namespace: "{{ .Release.Namespace }}",
			},
		},
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      traefik,
			Namespace: "{{ .Release.Namespace }}",
			Labels: map[string]string{
				appKubernetesName:     traefik,
				appKubernetesInstance: traefik + "-" + "{{ .Release.Namespace }}",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				appKubernetesName:     traefik,
				appKubernetesInstance: traefik + "-" + "{{ .Release.Namespace }}",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       traefik,
					Port:       int32(9000),
					TargetPort: intstr.Parse("traefik"),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       traefik,
					Port:       int32(80),
					TargetPort: intstr.Parse("web"),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "websecure",
					Port:       int32(443),
					TargetPort: intstr.Parse("websecure"),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	replicas := int32(1)
	maxSurge := intstr.Parse("1")
	maxUnavailable := intstr.Parse("0")
	terminationGracePeriodSeconds := int64(60)
	requestCPU, _ := resource.ParseQuantity("50m")
	requestMemory, _ := resource.ParseQuantity("200Mi")
	limitedCPU, _ := resource.ParseQuantity("1000m")
	limitedMemory, _ := resource.ParseQuantity("2Gi")
	readOnlyRootFilesystem := true
	runAsGroup := int64(0)
	runAsNonRoot := false
	runAsUser := int64(0)
	fsGroup := int64(65532)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      traefik,
			Namespace: "{{ .Release.Namespace }}",
			Labels: map[string]string{
				appKubernetesName:     traefik,
				appKubernetesInstance: traefik + "-" + "{{ .Release.Namespace }}",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					appKubernetesName:     traefik,
					appKubernetesInstance: traefik + "-" + "{{ .Release.Namespace }}",
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       &maxSurge,
					MaxUnavailable: &maxUnavailable,
				},
			},
			MinReadySeconds: int32(0),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/path":   "/metrics",
						"prometheus.io/port":   "9100",
					},
					Labels: map[string]string{
						appKubernetesName:     traefik,
						appKubernetesInstance: traefik + "-" + "{{ .Release.Namespace }}",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName:            traefik,
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					HostNetwork:                   false,
					Containers: []corev1.Container{
						{
							Name:            traefik,
							Image:           "traefik:v2.9.7",
							ImagePullPolicy: corev1.PullIfNotPresent,
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
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/ping",
										Port:   intstr.Parse("8080"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								FailureThreshold:    int32(1),
								InitialDelaySeconds: int32(2),
								PeriodSeconds:       int32(10),
								SuccessThreshold:    int32(1),
								TimeoutSeconds:      int32(2),
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/ping",
										Port:   intstr.Parse("8080"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								FailureThreshold:    int32(3),
								InitialDelaySeconds: int32(2),
								PeriodSeconds:       int32(10),
								SuccessThreshold:    int32(1),
								TimeoutSeconds:      int32(2),
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "metrics",
									ContainerPort: 9100,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          traefik,
									ContainerPort: 8080,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "web",
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "websecure",
									ContainerPort: 443,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add:  []corev1.Capability{"NET_BIND_SERVICE"},
									Drop: []corev1.Capability{"ALL"},
								},
								ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
								RunAsGroup:             &runAsGroup,
								RunAsNonRoot:           &runAsNonRoot,
								RunAsUser:              &runAsUser,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data",
								},
								{
									Name:      "tmp",
									MountPath: "/tmp",
								},
							},
							Args: []string{
								"--global.checknewversion",
								"--global.sendanonymoususage",
								"--entrypoints.metrics.address=:9100/tcp",
								"--entrypoints.traefik.address=:8080/tcp",
								"--entrypoints.web.address=:80/tcp",
								"--entrypoints.websecure.address=:443/tcp",
								"--api.dashboard=true",
								"--ping=true",
								"--metrics.prometheus=true",
								"--metrics.prometheus.entrypoint=metrics",
								"--providers.kubernetescrd",
								"--providers.kubernetesingress=true",
								"--providers.kubernetescrd.namespaces={{ .Release.Namespace }",
								"--providers.kubernetesingress.namespaces={{ .Release.Namespace }}",
								"--entrypoints.websecure.http.tls=true",
								"--entrypoints.websecure.http.tls.domains[0].main=olares.com",
								"--entrypoints.websecure.http.tls.domains[0].sans=*.olares.com",
								"--log.level=DEBUG",
								"--accesslog=true",
								"--accesslog.fields.defaultmode=keep",
								"--accesslog.fields.headers.defaultmode=drop",
								"--serversTransport.insecureSkipVerify=true",
								"--api.insecure=true",
								"--api.dashboard=true",
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: traefik,
								},
							},
						},
						{
							Name: "tmp",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup: &fsGroup,
					},
				},
			},
		},
	}

	at.traefik = &Traefik{
		sa:          sa,
		pvc:         pvc,
		role:        role,
		roleBinding: roleBinding,
		svc:         svc,
		deployment:  deployment,
	}
	return at
}

func (at *AppTemplate) WriteFile(cfg *CreateConfig, owner string) (err error) {
	path := utils.GetAppPath(owner, cfg.Name)
	if existDir(path) {
		return os.ErrExist
	}
	createPath := filepath.Join(path, "templates")
	err = os.MkdirAll(createPath, os.ModePerm)
	if err != nil {
		klog.Errorf("failed to mkdir path %s err=%v", createPath, err)
		return err
	}
	if at.appCfg != nil {
		yml, err := ToYaml(at.appCfg)
		if err != nil {
			klog.Errorf("failed to convert appCfg to yaml %v", err)
			return err
		}
		filename := filepath.Join(path, constants.AppCfgFileName)
		err = ioutil.WriteFile(filename, yml, 0644)
		if err != nil {
			klog.Errorf("failed to write file %s err=%v", filename, err)
			return err
		}
	}
	if at.chartMetadata != nil {
		yml, err := ToYaml(at.chartMetadata)
		if err != nil {
			klog.Errorf("failed to convert chart metadata to yaml %v", err)
			return err
		}
		filename := filepath.Join(path, "Chart.yaml")
		err = ioutil.WriteFile(filename, yml, 0644)
		if err != nil {
			klog.Errorf("failed to write file %s, err=%v", filename, err)
			return err
		}
	}

	if at.owner != nil {
		filename := filepath.Join(path, "owners")
		err = ioutil.WriteFile(filename, []byte{}, 0644)
		if err != nil {
			klog.Errorf("failed to write file %s, err=%v", filename, err)
			return err
		}
	}
	var yml []byte
	if at.deployment != nil {
		yml, err = ToYaml(at.deployment)
		if err != nil {
			klog.Errorf("failed to convert deployment to yaml %v", err)
			return err
		}

	}
	var sep = []byte("\n---\n")
	if at.service != nil {
		serviceYml, err := ToYaml(at.service)
		if err != nil {
			klog.Errorf("failed to convert service to yaml %v", err)
			return err
		}
		yml = append(yml, sep...)
		yml = append(yml, serviceYml...)
	}
	filename := filepath.Join(path, "templates", "deployment.yaml")
	err = ioutil.WriteFile(filename, yml, 0644)
	if err != nil {
		klog.Errorf("failed to write file %s, err=%v", filename, err)
		return err
	}
	filename = filepath.Join(path, "values.yaml")
	err = ioutil.WriteFile(filename, nil, 0644)
	if err != nil {
		klog.Errorf("failed to write file %s, err=%v", filename, err)
		return err
	}
	if cfg.Traefik {
		err = at.WriteTraefikFile(path)
	}

	return err
}

func (at *AppTemplate) WriteTraefikFile(path string) error {
	var sep = []byte("\n---\n")

	// ServiceAccount
	yml, err := ToYaml(at.traefik.sa)
	if err != nil {
		klog.Errorf("failed to convert traefik.sa to yaml %v", err)
		return err
	}
	source := []byte("# Source: traefik/templates/rbac/serviceaccount.yaml\n")
	yml = append(append(source, yml...), sep...)

	// PersistentVolumeClaim
	pvcYml, err := ToYaml(at.traefik.pvc)
	if err != nil {
		klog.Errorf("failed to convert traefik.pvc to yaml %v", err)
		return err
	}
	source = []byte("# Source: traefik/templates/pvc.yaml\n")
	yml = append(append(yml, source...), pvcYml...)
	yml = append(yml, sep...)

	// Role
	roleYml, err := ToYaml(at.traefik.role)
	if err != nil {
		klog.Errorf("failed to convert traefik.role to yaml %v", err)
		return err
	}
	source = []byte("# Source: traefik/templates/rbac/clusterrole.yaml\n")
	yml = append(append(yml, source...), roleYml...)
	yml = append(yml, sep...)

	// RoleBinding
	roleBindingYml, err := ToYaml(at.traefik.roleBinding)
	if err != nil {
		klog.Errorf("failed to convert traefik.roleBinding to yaml %v", err)
		return err
	}
	source = []byte("# Source: traefik/templates/rbac/clusterrolebinding.yaml\n")
	yml = append(append(yml, source...), roleBindingYml...)
	yml = append(yml, sep...)

	// Service
	svcYml, err := ToYaml(at.traefik.svc)
	if err != nil {
		klog.Errorf("failed to convert traefik.svc to yaml %v", err)
		return err
	}
	source = []byte("# Source: traefik/templates/service.yaml\n")
	yml = append(append(yml, source...), svcYml...)
	yml = append(yml, sep...)

	// Deployment
	deploymentYml, err := ToYaml(at.traefik.deployment)
	if err != nil {
		return err
	}
	source = []byte("# Source: traefik/templates/deployment.yaml\n")
	yml = append(append(yml, source...), deploymentYml...)

	err = ioutil.WriteFile(filepath.Join(path, "templates", "traefik.yaml"), yml, 0644)
	return err
}

func ToYaml(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte{}, err
	}
	yml, err := yaml.JSONToYAML(b)
	if err != nil {
		return []byte{}, err
	}
	return yml, err
}

func SetCreateConfigByPrompt() (*CreateConfig, error) {
	cfg := &CreateConfig{}
	yes, no := "yes", "no"
	options := []string{yes, no}

	promptAppName := survey.Input{Message: "please enter application name:"}
	err := survey.AskOne(&promptAppName, &cfg.Name, survey.WithValidator(survey.Required))
	if err != nil {
		return nil, err
	}

	promptSysDB := survey.Select{
		Message: "do you need system db?",
		Options: options,
		Default: no,
	}
	var systemDB string
	err = survey.AskOne(&promptSysDB, &systemDB)
	if err != nil {
		return nil, err
	}

	if systemDB == yes {
		cfg.SystemDB = true
	}
	if cfg.SystemDB {
		selectedDB := make([]string, 0)

		promptDB := survey.MultiSelect{
			Message: "pick databases",
			Options: []string{"Redis", "Mongodb", "PostgreSQL"},
		}
		err = survey.AskOne(&promptDB, &selectedDB)
		if err != nil {
			return nil, err
		}
		for _, db := range selectedDB {
			if db == "Redis" {
				cfg.Redis = true
			}
			if db == "Mongodb" {
				cfg.MongoDB = true
			}
			if db == "PostgreSQL" {
				cfg.PostgreSQL = true
			}
		}
	}

	promptSysCall := survey.Select{
		Message: "do you need any API from other Applications?",
		Options: options,
		Default: no,
	}
	var systemCall string
	err = survey.AskOne(&promptSysCall, &systemCall)
	if err != nil {
		return nil, err
	}
	if systemCall == yes {
		cfg.SystemCall = true
	}

	promptImage := survey.Input{
		Message: "please enter a mirror of the main program:",
	}
	err = survey.AskOne(&promptImage, &cfg.Img, survey.WithValidator(survey.Required))
	if err != nil {
		return nil, err
	}

	promptPort := survey.Input{
		Message: "which ports do you want to enable",
		Help:    "separated with comma,like 80,443",
	}
	var portStr string
	err = survey.AskOne(&promptPort, &portStr, survey.WithValidator(func(s interface{}) error {
		ports := strings.Split(s.(string), ",")
		for _, port := range ports {
			_, e := strconv.Atoi(port)
			if e != nil {
				return e
			}
		}
		return nil
	}))
	if err != nil {
		return nil, err
	}
	for _, p := range strings.Split(portStr, ",") {
		port, _ := strconv.Atoi(p)
		cfg.Ports = append(cfg.Ports, port)
	}

	var ingressRouter string
	promptIngress := survey.Select{
		Message: "do you need a IngressRouter inside your Application?",
		Options: options,
		Default: no,
	}
	err = survey.AskOne(&promptIngress, &ingressRouter)
	if err != nil {
		return nil, err
	}
	if ingressRouter == yes {
		cfg.IngressRouter = true
	}

	if cfg.IngressRouter {
		var ingress string
		promptRouter := survey.Select{
			Message: "pick a router",
			Options: []string{"Traefik"},
		}
		err = survey.AskOne(&promptRouter, &ingress)
		if err != nil {
			return nil, err
		}
		if ingress == "Traefik" {
			cfg.Traefik = true
		}
	} else {
		ports := make([]string, 0)
		for _, p := range cfg.Ports {
			ports = append(ports, strconv.Itoa(p))
		}
		if len(ports) > 1 {
			promptWebPort := survey.Select{
				Message: "pick a port for website",
				Options: ports,
			}
			err = survey.AskOne(&promptWebPort, &cfg.WebsitePort)
			if err != nil {
				return nil, err
			}
		} else {
			cfg.WebsitePort = strconv.Itoa(cfg.Ports[0])
		}
	}

	var appData string
	promptAppData := survey.Select{
		Message: "do you need something files on local only you can visit",
		Options: options,
		Default: no,
	}
	err = survey.AskOne(&promptAppData, &appData)
	if err != nil {
		return nil, err
	}
	if appData == yes {
		cfg.AppData = true
	}

	var userData string
	promptUserData := survey.Select{
		Message: "do you want visit users data",
		Options: options,
	}
	err = survey.AskOne(&promptUserData, &userData)
	if err != nil {
		return nil, err
	}
	if userData == yes {
		cfg.UserData = []string{}
	}

	var needGpu string
	promptGpu := survey.Select{
		Message: "do you need gpu",
		Options: options,
		Default: no,
	}
	err = survey.AskOne(&promptGpu, &needGpu)
	if err != nil {
		return nil, err
	}
	if needGpu == yes {
		cfg.NeedGPU = true
	}
	if cfg.NeedGPU {
		promptRequiredGPU := survey.Input{
			Message: "What size gpu do you need? (G)",
		}
		err = survey.AskOne(&promptRequiredGPU, &cfg.RequiredGPU)
		if err != nil {
			return nil, err
		}
	}

	promptRequiredMemory := survey.Input{
		Message: "How much memory your application needs? (G)",
	}
	err = survey.AskOne(&promptRequiredMemory, &cfg.RequiredMemory)
	if err != nil {
		return nil, err
	}

	var osVersion string
	promptOsVersion := survey.Input{
		Message: "The minimum system version you support? (eg: 0.3.0-0) or need a max version you can enter <min_version>,<max_version> (eg: 0.4.0-0)",
	}
	err = survey.AskOne(&promptOsVersion, &osVersion, survey.WithValidator(survey.Required), survey.WithValidator(func(obj interface{}) error {
		s := obj.(string)
		var version string
		versions := strings.Split(s, ",")
		version = ">= " + versions[0]
		if len(versions) > 1 {
			version = version + " <=" + versions[1]
		}
		_, e := semver.NewConstraint(version)
		return e
	}))
	if err != nil {
		return nil, err
	}
	versions := strings.Split(osVersion, ",")
	cfg.OSVersion = ">= " + versions[0]
	if len(versions) > 1 {
		cfg.OSVersion = cfg.OSVersion + " <=" + versions[1]
	}

	var specialOsVersion string
	promptOsVersion2 := survey.Input{
		Message: "Do you support some special version?",
	}
	err = survey.AskOne(&promptOsVersion2, &specialOsVersion, survey.WithValidator(func(obj interface{}) error {
		s := obj.(string)
		if len(s) == 0 {
			return nil
		}
		var version string
		for _, v := range strings.Split(s, ",") {
			if len(version) > 0 {
				version += "|| "
			}
			version = version + "=" + v
		}
		_, e := semver.NewConstraint(version)
		return e
	}))
	if err != nil {
		return nil, err
	}
	if len(specialOsVersion) != 0 {
		for _, v := range strings.Split(specialOsVersion, ",") {
			cfg.OSVersion = cfg.OSVersion + "|| =" + v
		}
	}

	printTable(cfg)
	return cfg, nil
}

func printTable(cfg *CreateConfig) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Field", "Value"})
	t.AppendRows([]table.Row{
		{"name", cfg.Name},
		{"systemDB", cfg.SystemDB},
		{"redis", cfg.Redis},
		{"mongodb", cfg.MongoDB},
		{"postgreSQL", cfg.PostgreSQL},
		{"systemCall", cfg.SystemCall},
		{"img", cfg.Img},
		{"ports", cfg.Ports},
		{"ingressRouter", cfg.IngressRouter},
		{"traefik", cfg.Traefik},
		{"appData", cfg.AppData},
		{"userData", cfg.UserData},
		{"needGpu", cfg.NeedGPU},

		{"requiredMemory", cfg.RequiredMemory},
		{"osVersion", cfg.OSVersion},
	})
	if len(cfg.WebsitePort) > 0 {
		t.AppendRows([]table.Row{
			{"websitePort", cfg.WebsitePort},
		})
	}
	if cfg.NeedGPU {
		t.AppendRows([]table.Row{
			{"requiredGpu", cfg.RequiredGPU},
		})
	}
	t.AppendSeparator()
	t.SetStyle(table.StyleLight)
	t.Render()
}

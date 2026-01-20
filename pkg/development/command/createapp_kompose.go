package command

import (
	"errors"
	"fmt"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/beclab/devbox/pkg/constants"
	"github.com/beclab/devbox/pkg/utils"
	"github.com/beclab/oachecker"

	"helm.sh/helm/v3/pkg/chart"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	kresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/restmapper"
	ctrl "sigs.k8s.io/controller-runtime"
)

type CreateFromDockerCompose struct {
	Title string `json:"title"`
	Type  string `json:"type"`
}

type KomposeFileOpts struct {
	Cfg          *CreateFromDockerCompose
	Resources    []runtime.Object
	Owner        string
	Name         string
	EntranceHost string
	EntrancePort int32
}

var requests = corev1.ResourceList{
	corev1.ResourceCPU:    kresource.MustParse("100m"),
	corev1.ResourceMemory: kresource.MustParse("128Mi"),
}

var limits = corev1.ResourceList{
	corev1.ResourceCPU:    kresource.MustParse("200m"),
	corev1.ResourceMemory: kresource.MustParse("512Mi"),
}

func writeManifest(opts *KomposeFileOpts, totalRequests, totalLimits corev1.ResourceList) error {
	appRef := make([]string, 0)
	configType := opts.Cfg.Type
	if configType == "" {
		configType = "app"
	}
	appcfg := oachecker.AppConfiguration{
		ConfigVersion: "0.8.0",
		ConfigType:    configType,
		Metadata: oachecker.AppMetaData{
			Name:        opts.Name,
			Icon:        defaultIcon,
			Description: fmt.Sprintf("app %s", opts.Name),
			AppID:       opts.Name,
			Version:     "0.0.1",
			Title:       opts.Cfg.Title,
			Categories:  []string{"dev"},
		},
		Spec: oachecker.AppSpec{
			RequiredMemory: func() string {
				q := totalRequests[corev1.ResourceMemory]
				return q.String()
			}(),
			RequiredCPU: func() string {
				q := totalRequests[corev1.ResourceCPU]
				return q.String()
			}(),
			RequiredDisk: "50Mi",
			LimitedMemory: func() string {
				q := totalLimits[corev1.ResourceMemory]
				return q.String()
			}(),
			LimitedCPU: func() string {
				q := totalLimits[corev1.ResourceCPU]
				return q.String()
			}(),
			VersionName: "0.0.1",
			SupportArch: []string{"amd64"},
		},
		Options: oachecker.Options{
			AppScope: &oachecker.AppScope{
				AppRef: appRef,
			},
		},
	}
	entrances := make([]oachecker.Entrance, 0)
	entrances = append(entrances, oachecker.Entrance{
		Name:       opts.Name,
		Host:       opts.EntranceHost,
		Port:       opts.EntrancePort,
		Title:      opts.Cfg.Title,
		Icon:       defaultIcon,
		AuthLevel:  "private",
		OpenMethod: "default",
	})

	appcfg.Entrances = entrances

	appPath := utils.GetAppPath(opts.Owner, opts.Name)
	if err := os.MkdirAll(appPath, os.ModePerm); err != nil {
		return err
	}
	yml, err := ToYaml(appcfg)

	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(appPath, constants.AppCfgFileName), yml, 0644)
	if err != nil {
		return err
	}
	metadata := chart.Metadata{
		APIVersion:  "v2",
		Name:        opts.Name,
		Description: "description",
		Type:        "application",
		Version:     "0.0.1",
		AppVersion:  "0.0.1",
	}
	yml, err = ToYaml(metadata)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(appPath, "Chart.yaml"), yml, 0644)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(appPath, "values.yaml"), nil, 0644)
	if err != nil {
		return err
	}

	return nil
}

func addResourcesRequirements(resource runtime.Object) {
	switch obj := resource.(type) {
	case *appsv1.Deployment:
		addResourcesToContainers(obj.Spec.Template.Spec.Containers, requests, limits)
	case *appsv1.StatefulSet:
		addResourcesToContainers(obj.Spec.Template.Spec.Containers, requests, limits)
	case *appsv1.DaemonSet:
		addResourcesToContainers(obj.Spec.Template.Spec.Containers, requests, limits)
	case *corev1.Pod:
		addResourcesToContainers(obj.Spec.Containers, requests, limits)
	}
}

func addResourcesToContainers(containers []corev1.Container, requests, limits corev1.ResourceList) {
	for i := range containers {
		container := &containers[i]
		if container.Resources.Requests == nil {
			container.Resources.Requests = make(corev1.ResourceList)
		}
		if container.Resources.Limits == nil {
			container.Resources.Limits = make(corev1.ResourceList)
		}

		// set requests
		for key, value := range requests {
			if _, exists := container.Resources.Requests[key]; !exists {
				container.Resources.Requests[key] = value
			}
		}
		// set limits
		for key, value := range limits {
			if _, exists := container.Resources.Limits[key]; !exists {
				container.Resources.Limits[key] = value
			}
		}
	}
}

func accumulateContainerResources(containers []corev1.Container, totalRequests, totalLimits corev1.ResourceList) {
	for i := range containers {
		c := containers[i]
		for key, value := range c.Resources.Requests {
			if existing, ok := totalRequests[key]; ok {
				existing.Add(value)
				totalRequests[key] = existing
			} else {
				totalRequests[key] = value.DeepCopy()
			}
		}
		for key, value := range c.Resources.Limits {
			if existing, ok := totalLimits[key]; ok {
				existing.Add(value)
				totalLimits[key] = existing
			} else {
				totalLimits[key] = value.DeepCopy()
			}
		}
	}
}

func WriteKomposeFile(opts *KomposeFileOpts) error {
	if opts == nil {
		return errors.New("nil kompose opts")
	}
	appPath := utils.GetAppPath(opts.Owner, opts.Name)
	//chartPath := filepath.Join(appPath, opts.Name)
	templatesDir := filepath.Join(appPath, "templates")
	if err := os.MkdirAll(templatesDir, os.ModePerm); err != nil {
		return err
	}

	totalRequests := corev1.ResourceList{corev1.ResourceCPU: kresource.MustParse("100m"), corev1.ResourceMemory: kresource.MustParse("100Mi")}
	totalLimits := corev1.ResourceList{corev1.ResourceCPU: kresource.MustParse("100m"), corev1.ResourceMemory: kresource.MustParse("100Mi")}

	hasSetEntrance := false
	// write each resource into chart templates and accumulate resource totals
	for i := range opts.Resources {
		resource := opts.Resources[i]
		addResourcesRequirements(resource)
		nsScoped, err := isNamespaceScoped(resource)
		if err != nil {
			klog.Errorf("failed to check resource namespace scoped: %v", err)
			return err
		}

		if nsScoped {
			if obj, ok := resource.(metav1.Object); ok {
				obj.SetNamespace("{{ .Release.Namespace }}")
			}
		}

		switch obj := resource.(type) {
		case *appsv1.Deployment:
			klog.Infof("dname: %s,ana: %v", obj.GetName(), obj.GetAnnotations())
			if obj.Annotations["olares.service.type"] == "Entrance" && !hasSetEntrance {
				klog.Errorf("set olares.service.type....")
				obj.SetName(opts.Name)
				hasSetEntrance = true
			}
			accumulateContainerResources(obj.Spec.Template.Spec.Containers, totalRequests, totalLimits)
		case *appsv1.StatefulSet:
			if obj.Annotations["olares.service.type"] == "Entrance" && !hasSetEntrance {
				obj.SetName(opts.Name)
				hasSetEntrance = true
			}
			accumulateContainerResources(obj.Spec.Template.Spec.Containers, totalRequests, totalLimits)
		case *appsv1.DaemonSet:
			accumulateContainerResources(obj.Spec.Template.Spec.Containers, totalRequests, totalLimits)
		case *corev1.Pod:
			accumulateContainerResources(obj.Spec.Containers, totalRequests, totalLimits)
		}

		yml, err := ToYaml(resource)
		if err != nil {
			return err
		}
		kind := strings.ToLower(resource.GetObjectKind().GroupVersionKind().Kind)
		name := resource.(metav1.Object).GetName()
		filename := filepath.Join(templatesDir, fmt.Sprintf("%s-%s.yaml", kind, name))
		if err := ioutil.WriteFile(filename, yml, 0644); err != nil {
			return err
		}
	}
	if _, ok := totalRequests[corev1.ResourceCPU]; !ok {
		totalRequests[corev1.ResourceCPU] = requests[corev1.ResourceCPU].DeepCopy()
	}
	if _, ok := totalRequests[corev1.ResourceMemory]; !ok {
		totalRequests[corev1.ResourceMemory] = requests[corev1.ResourceMemory].DeepCopy()
	}
	if _, ok := totalLimits[corev1.ResourceCPU]; !ok {
		totalLimits[corev1.ResourceCPU] = limits[corev1.ResourceCPU].DeepCopy()
	}
	if _, ok := totalLimits[corev1.ResourceMemory]; !ok {
		totalLimits[corev1.ResourceMemory] = limits[corev1.ResourceMemory].DeepCopy()
	}

	err := writeManifest(opts, totalRequests, totalLimits)
	if err != nil {
		return err
	}

	return nil
}

var (
	mapperOnce sync.Once
	mapper     meta.RESTMapper
	mapperErr  error
)

func getRESTMapper() (meta.RESTMapper, error) {
	mapperOnce.Do(func() {
		cfg, err := ctrl.GetConfig()
		if err != nil {
			mapperErr = err
			return
		}
		dc, err := discovery.NewDiscoveryClientForConfig(cfg)
		if err != nil {
			mapperErr = err
			return
		}
		mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	})
	return mapper, mapperErr
}

func isNamespaceScoped(resource runtime.Object) (bool, error) {
	gvk := resource.GetObjectKind().GroupVersionKind()

	rm, err := getRESTMapper()
	if err != nil {
		return false, err
	}
	mapping, err := rm.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return false, err
	}
	return mapping.Scope.Name() == meta.RESTScopeNameNamespace, nil
}

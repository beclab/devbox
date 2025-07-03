package helm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/devbox/pkg/appcfg"
	"github.com/beclab/devbox/pkg/constants"
	"github.com/beclab/devbox/pkg/utils"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeSchema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	yml "sigs.k8s.io/yaml"
)

func LoadChart(path string) (*chart.Chart, error) {
	if ok, err := chartutil.IsChartDir(path); err != nil {
		klog.Error("path validate error, ", err)
		return nil, err
	} else if !ok {
		klog.Error("path is not a valid helm chart dir")
		return nil, errors.New("path is not a valid helm chart dir")
	}

	chart, err := loader.Load(path)
	if err != nil {
		klog.Error("load chart error, ", err)
		return nil, err
	}

	return chart, nil
}

func GetChartVersion(chart *chart.Chart) (*semver.Version, error) {
	if chart.Metadata == nil {
		return nil, errors.New("metadata of chart is missing")
	}

	version, err := semver.NewVersion(chart.Metadata.Version)
	if err != nil {
		klog.Error("version format error, ", err)
		return nil, err
	}

	return version, nil
}

func UpdateChartVersion(chart *chart.Chart, name, path string, newVersion *semver.Version) error {
	chart.Metadata.Version = newVersion.String()

	fileBytes, err := json.Marshal(chart.Metadata)
	if err != nil {
		klog.Error("encoding chart metadata error, ", err)
		return err
	}

	fileData, err := yml.JSONToYAML(fileBytes)
	if err != nil {
		klog.Error("chart metadata json to yaml error, ", err)
		return err
	}

	chartYamlPath := filepath.Join(path, "Chart.yaml")
	err = os.WriteFile(chartYamlPath, fileData, 0644)
	if err != nil {
		klog.Error("write file chart.yaml error, ", err)
		return err
	}

	return nil
}

func UpdateChartName(chart *chart.Chart, name, path string) error {
	chart.Metadata.Name = name + "-dev"

	fileData, err := yaml.Marshal(chart.Metadata)
	if err != nil {
		klog.Error("encoding chart metadata error, ", err)
		return err
	}

	chartYaml := filepath.Join(path, "Chart.yaml")
	err = os.WriteFile(chartYaml, fileData, 0644)
	if err != nil {
		klog.Error("write file chart.yaml error, ", err)
		return err
	}

	return nil
}

func UpdateAppCfgVersion(owner, path string, version *semver.Version) error {
	appCfgYaml := filepath.Join(path, constants.AppCfgFileName)
	data, err := os.ReadFile(appCfgYaml)
	if err != nil {
		klog.Error("read app cfg error, ", err, ", ", appCfgYaml)
		return err
	}
	//var appCfg application.AppConfiguration
	//err = yaml.Unmarshal(data, &appCfg)

	appCfg, err := utils.GetAppConfig(owner, data)
	if err != nil {
		klog.Error("parse appcfg error, ", err)
		return err
	}
	if version != nil {
		appCfg.Metadata.Version = version.String()
	}
	data, err = yaml.Marshal(&appCfg)
	if err != nil {
		klog.Error("encode appcfg error, ", err)
		return err
	}
	err = os.WriteFile(appCfgYaml, data, 0644)
	if err != nil {
		klog.Error("write file OlaresManifest.yaml error, ", err)
		return err
	}

	return nil
}

func UpdateAppCfgName(owner, name, path string) error {
	appDevName := name + "-dev"
	appCfgYaml := filepath.Join(path, constants.AppCfgFileName)
	data, err := os.ReadFile(appCfgYaml)
	if err != nil {
		klog.Error("read app cfg error, ", err, ", ", appCfgYaml)
		return err
	}

	//var appCfg application.AppConfiguration
	//err = yaml.Unmarshal(data, &appCfg)
	appCfg, err := utils.GetAppConfig(owner, data)
	if err != nil {
		klog.Error("parse OlaresManifest.yaml error, ", err)
		return err
	}

	appCfg.Metadata.Name = appDevName
	appCfg.Metadata.AppID = appDevName
	appCfg.Metadata.Title = appDevName

	//if version != nil {
	//	appCfg.Metadata.Version = version.String()
	//}

	for i, e := range appCfg.Entrances {
		appCfg.Entrances[i].Title = e.Title + "-dev"
	}

	data, err = yaml.Marshal(&appCfg)
	if err != nil {
		klog.Error("encode appcfg error, ", err)
		return err
	}

	err = os.WriteFile(appCfgYaml, data, 0644)
	if err != nil {
		klog.Error("write file OlaresManifest.yaml error, ", err)
		return err
	}

	return nil
}

func UpgradeChartVersion(chart *chart.Chart, name, path string, version *semver.Version) error {
	return UpdateChartVersion(chart, name, path, version)
}

// try to render the helm chart, and return the rendered manifest or error
func DryRun(ctx context.Context, kubeConfig *rest.Config, namespace, app, path string, vals map[string]interface{}) (string, error) {
	settings := cli.New()
	settings.KubeAPIServer = kubeConfig.Host
	settings.KubeToken = kubeConfig.BearerToken
	settings.KubeInsecureSkipTLSVerify = true

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		klog.Error("init helm action config error, ", err)
		return "", err
	}

	install := action.NewInstall(actionConfig)
	install.DryRun = true
	install.Timeout = 300 * time.Second
	install.Namespace = namespace
	install.ReleaseName = app

	chart, err := LoadChart(path)
	if err != nil {
		klog.Error("load chart error, ", err, ", ", path)
		return "", err
	}

	r, err := install.RunWithContext(ctx, chart, vals)
	if err != nil {
		klog.Error("install chart dry run error, ", err)
		return "", err
	}

	logReleaseInfo(r)
	return r.Manifest, nil
}

func GetRelease(kubeConfig *rest.Config, namespace, app string) error {
	settings := cli.New()
	settings.KubeAPIServer = kubeConfig.Host
	settings.KubeToken = kubeConfig.BearerToken
	settings.KubeInsecureSkipTLSVerify = true

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		klog.Error("init helm action config error, ", err)
		return err
	}
	history := action.NewHistory(actionConfig)
	_, err := history.Run(app)
	return err
}

func Uninstall(ctx context.Context, kubeConfig *rest.Config, namespace, app string) error {
	settings := cli.New()
	settings.KubeAPIServer = kubeConfig.Host
	settings.KubeToken = kubeConfig.BearerToken
	settings.KubeInsecureSkipTLSVerify = true

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		klog.Error("init helm action config error, ", err)
		return err
	}

	//history := action.NewHistory(actionConfig)
	//if _, err := history.Run(app); err != nil {
	//	if err != driver.ErrReleaseNotFound {
	//		klog.Error("query app history error, ", err, ", ", app, ", ", namespace)
	//		return err
	//	}
	//
	//	// history not found, uninstall is unnecessary
	//	klog.Info("app history not found, ", app, ", ", namespace)
	//	return nil
	//}

	uninstall := action.NewUninstall(actionConfig)
	uninstall.Timeout = 300 * time.Second
	uninstall.KeepHistory = false

	res, err := uninstall.Run(app)
	if err != nil {
		klog.Error("uninstall app error, ", err, ", ", app, ", ", namespace)
		return err
	}

	logUninstallReleaseInfo(res)

	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		klog.Error(err)
		return err
	}

	err = client.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return wait.PollUntilContextTimeout(ctx, time.Second, 3*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		_, err = client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}

			return false, err
		}

		return false, nil
	})
}

func logReleaseInfo(release *release.Release) {
	klog.Info("app installed success, ",
		"NAME, ", release.Name,
		"LAST DEPLOYED, ", release.Info.LastDeployed.Format(time.ANSIC),
		"NAMESPACE, ", release.Namespace,
		"STATUS, ", release.Info.Status.String(),
		"REVISION, ", release.Version)
}

func logUninstallReleaseInfo(release *release.UninstallReleaseResponse) {
	klog.Info("app uninstalled success, ",
		"NAME, ", release.Release.Name,
		"NAMESPACE, ", release.Release.Namespace,
		"INFO, ", release.Info)
}

func DecodeManifest(manifest string) ([]runtime.Object, error) {
	var resources []runtime.Object
	decoder := yamlutil.NewYAMLOrJSONDecoder(strings.NewReader(manifest), 10)
	deserializer := scheme.Codecs.UniversalDeserializer()
	for {
		var rawObj runtime.RawExtension
		if err := decoder.Decode(&rawObj); err != nil {
			if err != io.EOF {
				klog.Warning("decode object error, ", err)
			}
			break
		}

		if len(strings.TrimSpace(string(rawObj.Raw))) > 0 {
			obj, _, err := deserializer.Decode(rawObj.Raw, nil, nil)
			if err != nil {
				klog.Warning("deserialize object error, ", err)
				continue
			}
			resources = append(resources, obj)
		}

	} // end loop of decode object

	return resources, nil
}

func FindContainers(objs []runtime.Object) []*ContainerInfo {
	var infos []*ContainerInfo

	getPodSelector := func(templ *corev1.PodTemplateSpec) string {
		var selectors []string
		for k, l := range templ.Labels {
			selectors = append(selectors, strings.Join([]string{k, l}, "="))
		}

		// sort to keep the expression stabilize
		sort.Slice(selectors, func(i, j int) bool {
			return strings.Compare(selectors[i], selectors[j]) < 0
		})

		return strings.Join(selectors, ",")
	}

	for _, o := range objs {
		var podTemplate *corev1.PodTemplateSpec
		switch obj := o.(type) {
		case *appsv1.Deployment:
			podTemplate = &obj.Spec.Template
		case *appsv1.StatefulSet:
			podTemplate = &obj.Spec.Template
		case *appsv1.DaemonSet:
			podTemplate = &obj.Spec.Template
		default:
			continue
		}

		pod := getPodSelector(podTemplate)
		for _, c := range podTemplate.Spec.Containers {
			info := &ContainerInfo{
				PodSelector:   pod,
				ContainerName: c.Name,
				Image:         c.Image,
			}

			infos = append(infos, info)
		}
	}

	return infos
}

func GetAppCfg(appManagerName string) (*appcfg.ApplicationConfig, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		klog.Errorf("failed to get kubeconfig %v", err)
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		klog.Errorf("failed to creat dynamic client %v", err)
		return nil, err
	}
	gvr := runtimeSchema.GroupVersionResource{
		Group:    "app.bytetrade.io",
		Version:  "v1alpha1",
		Resource: "applicationmanagers",
	}
	am, err := dynamicClient.Resource(gvr).Namespace("").Get(context.TODO(), appManagerName, metav1.GetOptions{})
	if am == nil || err != nil {
		klog.Errorf("failed to get app manager name=%s, err=%v", appManagerName, err)
		return nil, err
	}

	data, _, _ := unstructured.NestedString(am.Object, "spec", "config")
	var applicationConfig appcfg.ApplicationConfig
	err = json.Unmarshal([]byte(data), &applicationConfig)
	if err != nil {
		klog.Errorf("failed to unmarshal application manager config err=%v", err)
		return nil, err
	}
	return &applicationConfig, nil

}

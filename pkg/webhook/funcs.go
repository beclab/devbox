package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/beclab/devbox/pkg/constants"
	"github.com/beclab/devbox/pkg/development/container"
	"github.com/beclab/devbox/pkg/development/envoy"
	"github.com/beclab/devbox/pkg/development/helm"
	"github.com/beclab/devbox/pkg/store/db"
	"github.com/beclab/devbox/pkg/store/db/model"

	"github.com/containerd/containerd/reference/docker"
	"github.com/google/uuid"
	"gorm.io/gorm"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (wh *Webhook) PatchAdmissionResponse(resp *admissionv1.AdmissionResponse, patchBytes []byte) {
	resp.Patch = patchBytes
	pt := admissionv1.PatchTypeJSONPatch
	resp.PatchType = &pt
}

// AdmissionError wraps error as AdmissionResponse
func (wh *Webhook) AdmissionError(err error) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

// mutate the developing app's name to "<app name>-dev"
func (wh *Webhook) MutateAppName(ctx context.Context, req *admissionv1.AdmissionRequest) (patch []byte, err error) {
	raw := req.Object.Raw
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(raw, nil, nil)
	if err != nil {
		klog.Error("deserialize object error, ", err)
		return nil, err
	}

	klog.Info("start to mutate workload kind, ", obj.GetObjectKind())
	switch workload := obj.(type) {
	case *appsv1.Deployment:
		return mutateName[*deployment](ctx, wh, (*deployment)(workload), raw, req.Operation)
	case *appsv1.StatefulSet:
		return mutateName[*statefulset](ctx, wh, (*statefulset)(workload), raw, req.Operation)
	case *appsv1.DaemonSet:
		return mutateName[*daemonset](ctx, wh, (*daemonset)(workload), raw, req.Operation)
	}

	return nil, nil
}

// check the deployment is the app main workload or not
// just the app main workload must to be mutated
func (wh *Webhook) mustMutateApp(ctx context.Context, releaseName string) (bool, error) {
	var devApps []*model.DevApp
	if err := wh.DB.DB.Where("app_name = ?", appName(releaseName)).Find(&devApps).Error; err != nil {
		klog.Error("exec sql error, ", err)
		return false, err
	}

	if len(devApps) == 0 {
		return false, nil
	}

	return true, nil
}

func devName(workloadName string, releaseName string) string {
	if workloadName == releaseName {
		return workloadName
	}
	return workloadName + "-dev"
}

func appName(releaseName string) string {
	return strings.TrimSuffix(releaseName, "-dev")
}

func makePatches[T any](original []byte, obj T, name string) ([]byte, error) {
	current, err := json.Marshal(obj)
	if err != nil {
		klog.Error("Error marshaling object, ", name)
		return nil, err
	}
	admissionResponse := admission.PatchResponseFromRaw(original, current)
	return json.Marshal(admissionResponse.Patches)
}

type workloadInterface interface {
	GetObjectMeta() *metav1.ObjectMeta
	GetPodTemplate() *corev1.PodTemplateSpec
}

type deployment appsv1.Deployment

func (d *deployment) GetObjectMeta() *metav1.ObjectMeta       { return &d.ObjectMeta }
func (d *deployment) GetPodTemplate() *corev1.PodTemplateSpec { return &d.Spec.Template }

type statefulset appsv1.StatefulSet

func (s *statefulset) GetObjectMeta() *metav1.ObjectMeta       { return &s.ObjectMeta }
func (s *statefulset) GetPodTemplate() *corev1.PodTemplateSpec { return &s.Spec.Template }

type daemonset appsv1.DaemonSet

func (d *daemonset) GetObjectMeta() *metav1.ObjectMeta       { return &d.ObjectMeta }
func (d *daemonset) GetPodTemplate() *corev1.PodTemplateSpec { return &d.Spec.Template }

func mutateName[T workloadInterface](ctx context.Context, wh *Webhook, workload T, raw []byte, op admissionv1.Operation) (patch []byte, err error) {
	workloadName := workload.GetObjectMeta().Name

	// helm release name is <appname>-dev
	releaseName := workload.GetObjectMeta().GetAnnotations()[helmRelease]

	// helm release namespace is <appname>-dev-<owner>
	releaseNamespace := workload.GetObjectMeta().GetAnnotations()[helmReleaseNamespace]

	klog.Info("start to mutate workload name if necessary, ", workloadName)
	ok, err := wh.mustMutateApp(ctx, releaseName)
	if err != nil {
		return nil, err
	}

	if !ok {
		klog.Info("not a developing app, ignore workload name , ", workloadName)
		return nil, nil
	}

	// TODO: sys-app's namespace is user-space-<owner>
	workloadDevName := devName(workloadName, releaseName)

	if op == admissionv1.Create {
		klog.Info("mutate workload name, ", workloadName, " to ", workloadDevName)
		workload.GetObjectMeta().Name = workloadDevName
	}
	if workload.GetPodTemplate().Annotations == nil {
		workload.GetPodTemplate().Annotations = make(map[string]string)
	}
	workload.GetPodTemplate().Annotations[helmRelease] = releaseName
	workload.GetPodTemplate().Annotations[helmReleaseNamespace] = releaseNamespace

	var app *model.DevApp
	err = wh.DB.DB.Where("app_name = ?", appName(releaseName)).First(&app).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Error("exec sql error, ", err)
		return nil, err
	}

	if err == nil {
		containers := make([]*model.DevAppContainers, 0)
		if err = wh.DB.DB.Where("app_id = ?", app.ID).Find(&containers).Error; err != nil {
			klog.Error("exec sql error, ", err)
			return nil, err
		}

		if len(containers) > 0 {
			var ids []string
			for _, c := range containers {
				ids = append(ids, strconv.Itoa(int(c.ContainerID)))
			}

			sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

			klog.Info("add the annotation of dev containers to pods, ", ids)
			workload.GetPodTemplate().Annotations[devContainers] = strings.Join(ids, ",")
		}
	}

	return makePatches(raw, workload, workloadName)
}

// mutate the pod in a developing app which has some containers that need to be replaced with a dev-container
func (wh *Webhook) MutatePodContainers(ctx context.Context, namespace string, raw []byte, proxyUUID uuid.UUID, baseDir string) (patch []byte, err error) {
	var pod corev1.Pod
	if err := json.Unmarshal(raw, &pod); err != nil {
		klog.Errorf("Error unmarshaling request to pod, ", err)
		return nil, err
	}

	app, matches, err := wh.mustMutatePod(ctx, &pod)
	if err != nil {
		klog.Error("Error checking pod, ", err)
		return nil, err
	}

	if len(matches) == 0 {
		return nil, nil
	}

	var endpoints []*envoy.DevcontainerEndpoint
	devPort := 5000
	firstMutateContainer := true
	for _, m := range matches {
		ep, err := wh.mutateContainerToDevContainer(ctx, &pod, m, devPort, firstMutateContainer)
		if err != nil {
			return nil, err
		}

		if ep != nil {
			endpoints = append(endpoints, ep)
			devPort++
			firstMutateContainer = false
		}
	}

	if len(endpoints) > 0 {
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		pod.Annotations[envoy.UUIDAnnotation] = proxyUUID.String()

		realapp := strings.TrimSuffix(app, "-dev")
		appcfg, err := helm.GetAppCfg(realapp, baseDir)
		if err != nil {
			return nil, err
		}

		err = envoy.InjectSidecar(ctx, wh.KubeClient, namespace, &pod, endpoints, proxyUUID.String(), appcfg)
		if err != nil {
			return nil, err
		}
	}
	return makePatches(raw, &pod, pod.Name)
}

func (wh *Webhook) mustMutatePod(ctx context.Context, pod *corev1.Pod) (string, []*model.DevAppContainers, error) {
	releaseName, ok := pod.Annotations[helmRelease]
	if !ok {
		return "", nil, nil
	}

	klog.Info("try to find release, ", releaseName)

	var app *model.DevApp
	err := wh.DB.DB.Where("app_name = ?", appName(releaseName)).First(&app).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Error("exec sql error, ", err)
		return "", nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil, nil
	}

	klog.Info("try to find app bind containers, ", app.ID)

	containers := make([]*model.DevAppContainers, 0)
	err = wh.DB.DB.Where("app_id = ?", app.ID).Find(&containers).Error
	if err != nil {
		klog.Error("exec sql error, ", err)
		return "", nil, err
	}

	if len(containers) == 0 {
		return "", nil, nil
	}

	matches := make([]*model.DevAppContainers, 0)
	for _, c := range containers {
		selector, err := labels.Parse(c.PodSelector)
		if err != nil {
			klog.Error("containers in dev_app_containers has an invalid pod selector, ", err, ", ", c.PodSelector, ", ", c.ID)
			return "", nil, err
		}

		klog.Info("try to match pod selector, ", c.PodSelector)
		if selector.Matches(labels.Set(pod.Labels)) {
			matches = append(matches, c)
		}
	}

	return releaseName, matches, nil
}

func (wh *Webhook) mutateContainerToDevContainer(ctx context.Context, pod *corev1.Pod, devcontainer *model.DevAppContainers, devPort int, firstMutateContainer bool) (*envoy.DevcontainerEndpoint, error) {
	for i, c := range pod.Spec.Containers {
		if c.Name == devcontainer.ContainerName {
			klog.Info("mutating container, ", c.Name, ", ", pod.Name, ", ", pod.Namespace)
			// change container image to dev image
			var dc *model.DevContainers
			err := wh.DB.DB.Where("id = ?", devcontainer.ContainerID).First(&dc).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				klog.Error("exec sql error, ", err)
				return nil, err
			}

			if errors.Is(err, gorm.ErrRecordNotFound) {
				klog.Error("container not found, ", devcontainer.ContainerID)
				return nil, errors.New("container not found")
			}

			pod.Spec.Containers[i].Image = container.DevEnvImage(dc.DevEnv)

			// make sure only one nginx is running in the pod
			if firstMutateContainer {
				// start code-server on custom port
				pod.Spec.Containers[i].Command = []string{
					"sh",
					"-c",
					`if [ ! -f /etc/nginx/conf.d/dev/dev.conf ]; then cp /etc/nginx/conf.d/dev.example /etc/nginx/conf.d/dev/dev.conf;fi;
				nginx && 
				exec /usr/bin/code-server --bind-addr "0.0.0.0:` + strconv.Itoa(devPort) + `" --auth=none --log=debug`,
				}
			} else {
				pod.Spec.Containers[i].Command = []string{
					"sh",
					"-c",
					`if [ ! -f /etc/nginx/conf.d/dev/dev.conf ]; then cp /etc/nginx/conf.d/dev.example /etc/nginx/conf.d/dev/dev.conf;fi;
				exec /usr/bin/code-server --bind-addr "0.0.0.0:` + strconv.Itoa(devPort) + `" --auth=none --log=debug`,
				}
			}

			endpoint := &envoy.DevcontainerEndpoint{
				Host: "localhost",
				Port: devPort,
				Name: pod.Spec.Containers[i].Name,
				Path: "/proxy/" + strconv.Itoa(devPort) + "/",
			}

			addToEnv := func(key, value string) {
				found := false
				for index, env := range pod.Spec.Containers[i].Env {
					if env.Name == key {
						pod.Spec.Containers[i].Env[index].Value = value
						found = true
					}
				}

				if !found {
					pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, corev1.EnvVar{
						Name:  key,
						Value: value,
					})
				}
			}

			// add container id to env
			addToEnv(container.DevContainerEnv, strconv.Itoa(int(dc.ID)))

			// add container port to env
			addToEnv(container.DevContainerPortEnv, strconv.Itoa(devPort))

			// volume mount for dev container
			volumes := pod.Spec.Volumes
			volumeMounts := pod.Spec.Containers[i].VolumeMounts

			// clear prev volume defines
			var newVols []corev1.Volume
			for _, v := range volumes {
				switch v.Name {
				case "user-cache-dir":
					continue
				}

				newVols = append(newVols, v)
			}

			volumes = newVols

			var newVolMnts []corev1.VolumeMount
			for _, vm := range volumeMounts {
				switch vm.Name {
				case "user-cache-dir":
					continue
				}

				newVolMnts = append(newVolMnts, vm)
			}

			volumeMounts = newVolMnts
			directoryOrCreateType := corev1.HostPathDirectoryOrCreate

			userCacheDir, err := wh.getUserCacheDir(ctx)
			if err != nil {
				return nil, err
			}
			volumeName := "user-cache-dir"
			volumes = append(volumes, corev1.Volume{
				Name: volumeName,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Type: &directoryOrCreateType,
						Path: filepath.Join(userCacheDir, "studio", devcontainer.AppName),
					},
				},
			})
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      volumeName,
				MountPath: "/root",
			})

			pod.Spec.Volumes = volumes
			pod.Spec.Containers[i].VolumeMounts = volumeMounts

			klog.Info("bound devcontainer to pod")
			return endpoint, nil
		}
	}

	return nil, nil
}

func (wh *Webhook) getUserspaceDir(ctx context.Context) (string, error) {
	namespace := "user-space-" + constants.Owner
	bfl, err := wh.KubeClient.AppsV1().StatefulSets(namespace).Get(ctx, "bfl", metav1.GetOptions{})
	if err != nil {
		klog.Error("get user's bfl error, ", err)
		return "", err
	}

	dir, ok := bfl.Annotations["userspace_hostpath"]
	if !ok {
		klog.Error("user's space not found, ", err)
		return "", errors.New("userspace not found")
	}

	return dir, nil
}

func (wh *Webhook) getUserCacheDir(ctx context.Context) (string, error) {
	namespace := "user-space-" + constants.Owner
	bfl, err := wh.KubeClient.AppsV1().StatefulSets(namespace).Get(ctx, "bfl", metav1.GetOptions{})
	if err != nil {
		klog.Error("get user's bfl error, ", err)
		return "", err
	}
	dir, ok := bfl.Annotations["appcache_hostpath"]
	if !ok {
		klog.Error("user's cache dir not found, ", err)
		return "", errors.New("user cache dir not found")
	}
	return dir, nil
}

func (wh *Webhook) getUserApplicationDir(ctx context.Context) (string, error) {
	dir, err := wh.getUserspaceDir(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "Application"), nil
}

func (wh *Webhook) getUserHomeDir(ctx context.Context) (string, error) {
	dir, err := wh.getUserspaceDir(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "Home"), nil
}

func (wh *Webhook) MutateIm(ctx context.Context, raw []byte, proxyUUID uuid.UUID) (patch []byte, err error) {
	var obj unstructured.Unstructured
	if err := json.Unmarshal(raw, &obj); err != nil {
		klog.Errorf("Error unmarshaling request to unstructured err=%v", err)
		return nil, err
	}
	appName, _, _ := unstructured.NestedString(obj.Object, "spec", "appName")
	originAppName := strings.TrimSuffix(appName, "-dev")
	refs, _, _ := unstructured.NestedSlice(obj.Object, "spec", "refs")

	sql := `select dc.id,dc.dev_env,dc.name, dc.create_time, dc.update_time, ac.pod_selector,ac.app_id, ac.container_name,ac.image,a.app_name
	from dev_apps a
	join dev_app_containers ac on a.id = ac.app_id
	join dev_containers dc on ac.container_id = dc.id
	where a.app_name = '%s'`
	sql = fmt.Sprintf(sql, originAppName)
	list := make([]*model.DevContainerInfo, 0)
	db := db.NewDbOperator()
	err = db.DB.Raw(sql).Scan(&list).Error
	if err != nil {
		return nil, err
	}
	klog.Infof("len(list)=%v", len(list))
	if len(list) == 0 {
		return makePatches(raw, obj.Object, appName)
	}

	newRefs := make([]interface{}, 0)
	for _, r := range refs {
		name := r.(map[string]interface{})["name"].(string)
		pullPolicy := r.(map[string]interface{})["imagePullPolicy"].(string)
		image, _ := docker.ParseDockerRef(*list[0].Image)

		devImage, _ := docker.ParseDockerRef(container.DevEnvImage(list[0].DevEnv))
		if name == image.String() {
			name = devImage.String()
		}
		newRefs = append(newRefs, map[string]interface{}{
			"name":            name,
			"imagePullPolicy": pullPolicy,
		})
	}
	klog.Infof("newRefs: %#v", newRefs)
	err = unstructured.SetNestedSlice(obj.Object, newRefs, "spec", "refs")
	if err != nil {
		return nil, err
	}

	patch, err = makePatches(raw, obj.Object, appName)
	if err != nil {
		klog.Infof("make Patches err=%v", err)
	}
	klog.Infof("pathc: %v", string(patch))
	return patch, err
}

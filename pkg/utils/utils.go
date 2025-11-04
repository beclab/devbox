package utils

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/beclab/devbox/pkg/appcfg"
	"github.com/beclab/oachecker"
	"github.com/containerd/containerd/reference/docker"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeSchema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
)

func GetAdminUsername(ctx context.Context) (string, error) {
	gvr := schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}
	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		klog.Errorf("failed to get kube config %v", err)
		return "", err
	}
	client, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		klog.Errorf("failed get get client %v", err)
		return "", err
	}
	data, err := client.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get user list err=%v", err)
		return "", err
	}

	var admin string
	for _, u := range data.Items {
		if u.Object == nil {
			continue
		}
		annotations := u.GetAnnotations()
		ownerRole := annotations["bytetrade.io/owner-role"]
		if ownerRole == "owner" || ownerRole == "admin" {
			admin = u.GetName()
			break
		}
	}

	return admin, nil
}

func GetAppConfig(owner string, data []byte) (*oachecker.AppConfiguration, error) {
	admin, err := GetAdminUsername(context.TODO())
	if err != nil {
		klog.Errorf("failed to get admin %v", err)
		return nil, err
	}
	isAdmin, err := IsAdmin(context.TODO(), owner)
	if err != nil {
		klog.Errorf("failed to check user %s is admin %v", owner, err)
		return nil, err
	}

	opts := []func(map[string]interface{}){
		oachecker.WithAdmin(admin),
		oachecker.WithOwner(owner),
		WithIsAdmin(isAdmin),
	}
	appcfg, err := oachecker.GetAppConfigurationFromContent(data, opts...)
	return appcfg, nil
}
func WithIsAdmin(isAdmin bool) func(map[string]interface{}) {
	return func(values map[string]interface{}) {
		values["isAdmin"] = isAdmin

	}
}

// GetAdminUserList returns admin list, an error if there is any.
func GetAdminUserList(ctx context.Context) ([]string, error) {
	adminUserList := make([]string, 0)

	gvr := schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}
	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		klog.Errorf("failed to get kube config %v", err)
		return adminUserList, err
	}
	client, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		klog.Errorf("failed to new kube client %v", err)
		return adminUserList, err
	}
	data, err := client.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get user list err=%v", err)
		return adminUserList, err
	}

	for _, u := range data.Items {
		if u.Object == nil {
			continue
		}
		annotations := u.GetAnnotations()
		role := annotations["bytetrade.io/owner-role"]
		if role == "owner" || role == "admin" {
			adminUserList = append(adminUserList, u.GetName())
		}
	}

	return adminUserList, nil
}

func IsAdmin(ctx context.Context, owner string) (bool, error) {
	adminList, err := GetAdminUserList(ctx)
	if err != nil {
		klog.Errorf("failed to get admin user list %v", err)
		return false, err
	}
	for _, user := range adminList {
		if user == owner {
		}
		return true, nil
	}
	return false, nil
}

func GetDefaultHelloImage() string {
	helloImage := "beclab/studio-app:1.0.0"
	envHelloImage := os.Getenv("OLARES_STUDIO_HELLO_IMAGE")
	_, err := docker.ParseDockerRef(envHelloImage)
	if err != nil {
		return helloImage
	}
	return envHelloImage
}

func DevName(name string) string {
	return fmt.Sprintf("%s-dev", name)
}

func GetAppID(appName string) string {
	hash := md5.Sum([]byte(appName))
	hashString := hex.EncodeToString(hash[:])
	return hashString[:8]
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

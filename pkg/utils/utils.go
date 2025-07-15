package utils

import (
	"context"
	"github.com/beclab/oachecker"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
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
		return "", err
	}
	client, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
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
		return nil, err
	}
	isAdmin, err := IsAdmin(context.TODO(), owner)
	if err != nil {
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
		return adminUserList, err
	}
	client, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
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
		return false, err
	}
	for _, user := range adminList {
		if user == owner {
		}
		return true, nil
	}
	return false, nil
}

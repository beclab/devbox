package utils

import (
	"context"

	"github.com/beclab/devbox/pkg/constants"
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
		if annotations["bytetrade.io/owner-role"] == "platform-admin" {
			admin = u.GetName()
			break
		}
	}

	return admin, nil
}

func GetAppConfig(data []byte) (*oachecker.AppConfiguration, error) {
	admin, err := GetAdminUsername(context.TODO())
	if err != nil {
		return nil, err
	}
	opts := []func(map[string]interface{}){
		oachecker.WithAdmin(admin),
		oachecker.WithOwner(constants.Owner),
	}
	appcfg, err := oachecker.GetAppConfigurationFromContent(data, opts...)
	return appcfg, nil
}

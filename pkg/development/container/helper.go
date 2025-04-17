package container

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/beclab/devbox/pkg/constants"
	"github.com/beclab/devbox/pkg/store/db/model"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

const (
	UnknownStatus       = "unknown"
	DevContainerEnv     = "DEV_CONTAINER"
	DevContainerPortEnv = "DEV_CONTAINER_PORT"
)

// DevEnvImage return env image
func DevEnvImage(env string) string {
	switch env {
	case "NodeJS":
		return "beclab/node-ts-dev:0.1.1"
	case "Golang":
		return "beclab/go-dev:0.1.1"
	case "Python":
		return "beclab/python-dev:0.1.1"
	case "default":
		return "beclab/node-ts-dev:0.1.1"
	}

	return env
}

func IsSysAppDevImage(image string) bool {
	switch {
	case strings.Contains(image, "node-ts-dev"),
		strings.Contains(image, "go-dev"),
		strings.Contains(image, "python-dev"):

		return true
	}

	return false
}

func GetContainerStatus(ctx context.Context, kubeconfig *rest.Config, container *model.DevContainerInfo) (string, string, error) {
	client, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		klog.Error("get kubernetes client error, ", err)
		return UnknownStatus, "", err
	}

	namespace := *container.AppName + "-dev-" + constants.Owner
	userspace := "user-space-" + constants.Owner
	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		LabelSelector: *container.PodSelector,
	})

	if err != nil {
		klog.Error("find pods of container error, ", err)
		return UnknownStatus, "", err
	}

	for _, p := range pods.Items {
		if p.Namespace == namespace || p.Namespace == userspace {
			for _, c := range p.Status.ContainerStatuses {
				if c.Name == *container.ContainerName {
					for _, con := range p.Spec.Containers {
						if con.Name == c.Name {
							state := UnknownStatus
							port := ""
							for _, e := range con.Env {
								switch {
								case e.Name == DevContainerEnv && e.Value == strconv.Itoa(int(container.ID)):
									switch {
									case c.State.Waiting != nil:
										state = "Waiting"
									case c.State.Running != nil:
										state = "Running"
									case c.State.Terminated != nil:
										state = "Terminated"
									}

								case e.Name == DevContainerPortEnv:
									port = e.Value
								}
							}

							if state != UnknownStatus {
								return state, port, nil
							}
						}
					}
				}
			}
		}
	}

	klog.Error("container not found, ", *container)
	return UnknownStatus, "", errors.New("container not found")
}

func CreateOrUpdateDevNamespace(ctx context.Context, kubeconfig *rest.Config, app string) (string, error) {
	client, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		klog.Error("get kubernetes client error, ", err)
		return "", err
	}

	namespaceName := app + "-" + constants.Owner

	ns, err := client.CoreV1().Namespaces().Get(ctx, namespaceName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.Error("get namespace error, ", err, ", ", namespaceName)
			return "", err
		} else {
			namespace := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespaceName,
					Labels: map[string]string{
						constants.DevOwnerLabel: constants.Owner,
					},
				},
			}

			_, err = client.CoreV1().Namespaces().Create(ctx, &namespace, metav1.CreateOptions{})
			if err != nil {
				klog.Error("create dev namespace error, ", err, ", ", namespaceName)
				return "", err
			}
		}
	} else {
		retry.RetryOnConflict(retry.DefaultRetry, func() error {
			ns.Labels[constants.DevOwnerLabel] = constants.Owner
			_, err = client.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{})
			return err
		})
	}

	return namespaceName, nil
}

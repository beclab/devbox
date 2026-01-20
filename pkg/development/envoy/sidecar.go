package envoy

import (
	"context"
	"fmt"
	"strconv"

	"github.com/beclab/devbox/pkg/appcfg"
	"github.com/beclab/oachecker"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

func InjectSidecar(ctx context.Context, kubeClient *kubernetes.Clientset, namespace string,
	pod *corev1.Pod, devcontainers []*DevcontainerEndpoint, proxyUUID string, appConfig *appcfg.ApplicationConfig) error {
	injected, _ := IsInjectedPod(pod)
	sidecarConfig := &ConfigBuilder{
		owner: appConfig.OwnerName,
	}
	sidecarConfig.WithDevcontainers(devcontainers)
	if injected {
		klog.Info("envoy sidecar injected pod")
		if IsWebsocketEnabled(pod) {
			klog.Info("websocket sidecar injected pod")
			sidecarConfig.WithWebsocket()
		}
	} else if appConfig.WsConfig.URL != "" {
		sidecarConfig.WithWebsocket()
	}

	config, err := sidecarConfig.Build()
	if err != nil {
		klog.Error("build sidecar config error, ", err)
		return err
	}

	configMapName, err := createSidecarConfigMap(ctx, kubeClient, proxyUUID, namespace, config)
	if err != nil {
		klog.Error("create sidecar config map error, ", err, ", ", namespace)
		return err
	}

	if injected {
		// sidecar injected, ( app-service or devbox, whatever )
		// change the envoy config to this config
		for i, v := range pod.Spec.Volumes {
			if v.Name == SidecarConfigMapVolumeName {
				pod.Spec.Volumes[i].VolumeSource.ConfigMap.LocalObjectReference.Name = configMapName
			}
		}
	} else {
		// inject envoy sidecar and websocket sidecar if necessary
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
			Name: SidecarConfigMapVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
					Items: []corev1.KeyToPath{
						{
							Key:  EnvoyConfigFileName,
							Path: EnvoyConfigFileName,
						},
					},
				},
			},
		})

		pod.Spec.InitContainers = append(pod.Spec.InitContainers, getInitContainerSpec())
		klog.Info("inject dev sidecar")
		pod.Spec.Containers = append(pod.Spec.Containers, getEnvoySidecarContainerSpec(pod))
		if sidecarConfig.Websocket() {
			klog.Info("inject websocket sidecar")
			pod.Spec.Containers = append(pod.Spec.Containers, getWebSocketSideCarContainerSpec(&appConfig.WsConfig))
		}
	}
	return nil
}

func createSidecarConfigMap(
	ctx context.Context, kubeClient *kubernetes.Clientset,
	proxyUUID, namespace, sidecarConfig string,
) (string, error) {
	configMapName := fmt.Sprintf("%s-%s", SidecarConfigMapVolumeName, proxyUUID)
	cm, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return "", err
	}

	newConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
		},
		Data: map[string]string{
			EnvoyConfigFileName: sidecarConfig,
		},
	}

	if err == nil {
		// configmap found
		cm.Data = newConfigMap.Data
		if _, err := kubeClient.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{}); err != nil {
			klog.Errorf("update dev sidecar configmap %s in namespace %s error, %v", configMapName, namespace, err)
			return "", err
		}
	} else {
		if _, err := kubeClient.CoreV1().ConfigMaps(namespace).Create(ctx, newConfigMap, metav1.CreateOptions{}); err != nil {
			klog.Errorf("create dev sidecar configmap %s in namespace %s error, %v", configMapName, namespace, err)
			return "", err
		}
	}

	return configMapName, nil
}

func getInitContainerSpec() corev1.Container {
	iptablesInitCommand := generateIptablesCommands()
	enablePrivilegedInitContainer := true

	return corev1.Container{
		Name:            SidecarInitContainerName,
		Image:           "beclab/init:v1.2.3",
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: &corev1.SecurityContext{
			Privileged: &enablePrivilegedInitContainer,
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{
					"NET_ADMIN",
				},
			},
			RunAsNonRoot: pointer.BoolPtr(false),
			// User ID 0 corresponds to root
			RunAsUser: pointer.Int64Ptr(0),
		},
		Command: []string{"/bin/sh"},
		Args: []string{
			"-c",
			iptablesInitCommand,
		},
		Env: []corev1.EnvVar{
			{
				Name: "POD_IP",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.podIP",
					},
				},
			},
		},
	}
}

func generateIptablesCommands() string {
	cmd := fmt.Sprintf(`iptables-restore --noflush <<EOF
# sidecar interception rules
*nat
:PROXY_IN_REDIRECT - [0:0]
:PROXY_INBOUND - [0:0]
-A PROXY_IN_REDIRECT -p tcp -j REDIRECT --to-port %d
-A PREROUTING -p tcp -j PROXY_INBOUND
-A PROXY_INBOUND -p tcp --dport %d -j RETURN
-A PROXY_INBOUND -p tcp --dport 22 -j RETURN
-A PROXY_INBOUND -p tcp -j PROXY_IN_REDIRECT
COMMIT
EOF
`,
		EnvoyInboundListenerPort,
		EnvoyAdminPort,
	)

	return cmd
}

func getWebSocketSideCarContainerSpec(wsConfig *oachecker.WsConfig) corev1.Container {
	return corev1.Container{
		Name:            WsContainerName,
		Image:           WsContainerImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/ws-gateway"},
		Env: []corev1.EnvVar{
			{
				Name:  "WS_PORT",
				Value: strconv.Itoa(wsConfig.Port),
			},
			{
				Name:  "WS_URL",
				Value: wsConfig.URL,
			},
		},
	}
}

func getEnvoySidecarContainerSpec(pod *corev1.Pod) corev1.Container {
	clusterID := fmt.Sprintf("%s.%s", pod.Spec.ServiceAccountName, pod.Namespace)

	return corev1.Container{
		Name:            EnvoyContainerName,
		Image:           EnvoyImageVersion,
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: pointer.BoolPtr(false),
			RunAsUser: func() *int64 {
				uid := EnvoyUID
				return &uid
			}(),
		},
		Ports: getEnvoyContainerPorts(),
		VolumeMounts: []corev1.VolumeMount{{
			Name:      SidecarConfigMapVolumeName,
			ReadOnly:  true,
			MountPath: EnvoyConfigFilePath + "/" + EnvoyConfigFileName,
			SubPath:   EnvoyConfigFileName,
		}},
		Command: []string{"envoy"},
		Args: []string{
			"--log-level", DefaultEnvoyLogLevel,
			"-c", EnvoyConfigFilePath + "/" + EnvoyConfigFileName,
			"--service-cluster", clusterID,
		},
		Env: []corev1.EnvVar{
			{
				Name: "POD_UID",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.uid",
					},
				},
			},
			{
				Name: "POD_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
			{
				Name: "POD_NAMESPACE",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
			{
				Name: "POD_IP",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
			{
				Name: "SERVICE_ACCOUNT",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "spec.serviceAccountName",
					},
				},
			},
		},
	}
}

func getEnvoyContainerPorts() []corev1.ContainerPort {
	containerPorts := []corev1.ContainerPort{
		{
			Name:          EnvoyInboundListenerPortName,
			ContainerPort: EnvoyInboundListenerPort,
		},
	}

	livenessPort := corev1.ContainerPort{
		// Name must be no more than 15 characters
		Name:          "liveness-port",
		ContainerPort: EnvoyLivenessProbePort,
	}
	containerPorts = append(containerPorts, livenessPort)

	return containerPorts
}

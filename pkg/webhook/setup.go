package webhook

import (
	"context"
	"os"

	"github.com/beclab/devbox/pkg/constants"
	"github.com/beclab/devbox/pkg/store/db"

	admissionregv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

const (
	defaultCaPath              = "/etc/certs/ca.crt"
	webhookServiceName         = "studio-server"
	devContainerWebhookCfgName = "devcontainer-mutate-webhooks"
	imageManagerWebhookCfgName = "imagemanager-mutate-webhooks"
	imageManagerWebhookPrefix  = "imagemanager-webhook"
	mutatingWebhookNamePrefix  = "devcontainer-webhook"
	helmRelease                = "meta.helm.sh/release-name"
	helmReleaseNamespace       = "meta.helm.sh/release-namespace"
	devContainers              = "dev.bytetrade.io/dev-containers"
)

var (
	webhookServiceNamespace       = &constants.Namespace
	webhookPath                   = "/webhook/devcontainer"
	WebhookPort             int32 = 8083

	imageManagerWebhookPath = "/webhook/imagemanager"
	// WebhookServerListenAddress       = webhookServiceName + ":" + strconv.Itoa(int(WebhookPort))

	// codecs is the codec factory used by the deserialzer
	codecs = serializer.NewCodecFactory(runtime.NewScheme())

	// Deserializer is used to decode the admission request body
	Deserializer = codecs.UniversalDeserializer()

	UUIDAnnotation = "studio.bytetrade.io/proxy-uuid"
)

type Webhook struct {
	KubeClient *kubernetes.Clientset
	DB         *db.DbOperator
}

func (wh *Webhook) CreateOrUpdateDevContainerMutatingWebhook() error {
	failurePolicy := admissionregv1.Fail
	matchPolicy := admissionregv1.Exact
	scoped := admissionregv1.NamespacedScope
	webhookTimeout := int32(30)

	mwhLabels := map[string]string{"velero.io/exclude-from-backup": "true"}
	caBundle, err := os.ReadFile(defaultCaPath)
	if err != nil {
		return err
	}
	mwh := admissionregv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:   devContainerWebhookCfgName,
			Labels: mwhLabels,
		},
		Webhooks: []admissionregv1.MutatingWebhook{},
	}

	devwhName := mutatingWebhookName()
	devwh := admissionregv1.MutatingWebhook{
		Name: devwhName,
		ClientConfig: admissionregv1.WebhookClientConfig{
			CABundle: caBundle,
			Service: &admissionregv1.ServiceReference{
				Namespace: *webhookServiceNamespace,
				Name:      webhookServiceName,
				Path:      &webhookPath,
				Port:      &WebhookPort,
			},
		},
		FailurePolicy: &failurePolicy,
		MatchPolicy:   &matchPolicy,
		Rules: []admissionregv1.RuleWithOperations{
			{
				Operations: []admissionregv1.OperationType{admissionregv1.Create, admissionregv1.Update},
				Rule: admissionregv1.Rule{
					APIGroups:   []string{"*"},
					APIVersions: []string{"*"},
					Resources:   []string{"pods", "deployments", "statefulsets", "daemonsets"},
					Scope:       &scoped,
				},
			},
		},
		NamespaceSelector: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      constants.DevOwnerLabel,
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{constants.Owner},
				},
			},
		},
		SideEffects: func() *admissionregv1.SideEffectClass {
			sideEffect := admissionregv1.SideEffectClassNoneOnDryRun
			return &sideEffect
		}(),
		TimeoutSeconds:          &webhookTimeout,
		AdmissionReviewVersions: []string{"v1"},
	}

	mwh.Webhooks = append(mwh.Webhooks, devwh)
	if _, err = wh.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.Background(), &mwh, metav1.CreateOptions{}); err != nil {
		// Webhook already exists, update the webhook in this scenario
		if apierrors.IsAlreadyExists(err) {
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existing, err := wh.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.Background(), mwh.Name, metav1.GetOptions{})
				if err != nil {
					klog.Error("Error getting MutatingWebhookConfiguration ", err)
					return err
				}

				found := false
				for i, w := range existing.Webhooks {
					if w.Name == devwh.Name {
						found = true
						existing.Webhooks[i] = devwh
						break
					}
				}

				if !found {
					existing.Webhooks = append(existing.Webhooks, devwh)
				}

				_, err = wh.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(context.Background(), existing, metav1.UpdateOptions{})
				if err != nil && !apierrors.IsConflict(err) {
					klog.Error("Error updating MutatingWebhookConfiguration ", err)
				}
				return err
			})

			if err != nil {
				klog.Error("Error updating MutatingWebhookConfiguration ", err)
				return err
			}

		} else {
			klog.Error("Error creating MutatingWebhookConfiguration ", err)
			return err
		}
	}
	klog.Infof("Finished creating MutatingWebhookConfiguration %s", devContainerWebhookCfgName)
	return nil
}

func (wh *Webhook) DeleteDevContainerMutatingWebhook() error {
	devwhName := mutatingWebhookName()
	existing, err := wh.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.Background(), devContainerWebhookCfgName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Info("webhook configuration not found, ", devContainerWebhookCfgName)
			return nil
		}

		return err
	}

	for i, w := range existing.Webhooks {
		if w.Name == devwhName {
			if len(existing.Webhooks) == 1 {
				err := wh.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(context.Background(), devContainerWebhookCfgName, metav1.DeleteOptions{})
				if err != nil {
					klog.Info("delete webhook configuration error, ", err)
					return err
				}
			} else {
				return retry.RetryOnConflict(retry.DefaultRetry, func() error {
					updating, err := wh.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.Background(), devContainerWebhookCfgName, metav1.GetOptions{})
					if err != nil {
						if apierrors.IsNotFound(err) {
							klog.Info("webhook configuration not found, ", devContainerWebhookCfgName)
							return nil
						}

						return err
					}
					updating.Webhooks = append(existing.Webhooks[:i], existing.Webhooks[i+1:]...)

					klog.Info("removing the webhook, ", devwhName)
					_, err = wh.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(context.Background(), updating, metav1.UpdateOptions{})
					if !apierrors.IsConflict(err) {
						klog.Error("Error updating MutatingWebhookConfiguration ", err)
					}
					return err
				})
			}

		}
	}

	klog.Info("success to clean the devbox webhook")
	return nil
}

func (wh *Webhook) CreateOrUpdateImageManagerMutatingWebhook() error {
	failurePolicy := admissionregv1.Fail
	matchPolicy := admissionregv1.Exact
	webhookTimeout := int32(30)

	caBundle, err := os.ReadFile(defaultCaPath)
	if err != nil {
		return err
	}

	mwhLabels := map[string]string{"velero.io/exclude-from-backup": "true"}
	mwh := admissionregv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:   imageManagerWebhookCfgName,
			Labels: mwhLabels,
		},
		Webhooks: []admissionregv1.MutatingWebhook{},
	}
	imwh := admissionregv1.MutatingWebhook{
		Name: imageManagerWebhookName(),
		ClientConfig: admissionregv1.WebhookClientConfig{
			CABundle: caBundle,
			Service: &admissionregv1.ServiceReference{
				Namespace: *webhookServiceNamespace,
				Name:      webhookServiceName,
				Path:      &imageManagerWebhookPath,
				Port:      &WebhookPort,
			},
		},
		FailurePolicy: &failurePolicy,
		MatchPolicy:   &matchPolicy,
		Rules: []admissionregv1.RuleWithOperations{
			{
				Operations: []admissionregv1.OperationType{admissionregv1.Create},
				Rule: admissionregv1.Rule{
					APIGroups:   []string{"app.bytetrade.io"},
					APIVersions: []string{"*"},
					Resources:   []string{"imagemanagers"},
				},
			},
		},
		ObjectSelector: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      constants.DevOwnerLabel,
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{constants.Owner},
				},
			},
		},
		SideEffects: func() *admissionregv1.SideEffectClass {
			sideEffect := admissionregv1.SideEffectClassNoneOnDryRun
			return &sideEffect
		}(),
		TimeoutSeconds:          &webhookTimeout,
		AdmissionReviewVersions: []string{"v1"},
	}
	mwh.Webhooks = append(mwh.Webhooks, imwh)
	if _, err = wh.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.TODO(), &mwh, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existing, err := wh.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.TODO(), mwh.Name, metav1.GetOptions{})
				if err != nil {
					klog.Error("Error getting MutatingWebhookConfiguration ", err)
					return err
				}
				found := false
				for i, w := range existing.Webhooks {
					if w.Name == imwh.Name {
						found = true
						existing.Webhooks[i] = imwh
						break
					}
				}
				if !found {
					existing.Webhooks = append(existing.Webhooks, imwh)
				}
				_, err = wh.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(context.TODO(), existing, metav1.UpdateOptions{})
				if err != nil && !apierrors.IsConflict(err) {
					klog.Error("Error updating MutatingWebhookConfiguration ", err)
				}
				return err
			})
			if err != nil {
				klog.Error("Error updating MutatingWebhookConfiguration ", err)
				return err
			}
		} else {
			klog.Error("Error creating MutatingWebhookConfiguration ", err)
			return err
		}
	}
	klog.Infof("Finished creating MutatingWebhookConfiguration %s", imageManagerWebhookCfgName)
	return nil
}

func (wh *Webhook) DeleteImageManagerMutatingWebhook() error {
	imwhName := imageManagerWebhookName()
	existing, err := wh.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.TODO(), imageManagerWebhookCfgName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Info("webhook configuration not found, ", imageManagerWebhookCfgName)
			return nil
		}
		return err
	}
	for i, w := range existing.Webhooks {
		if w.Name == imwhName {
			if len(existing.Webhooks) == 1 {
				err = wh.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(context.TODO(), imageManagerWebhookCfgName, metav1.DeleteOptions{})
				if err != nil {
					klog.Info("delete webhook configuration error, ", err)
					return err
				}
			} else {
				return retry.RetryOnConflict(retry.DefaultRetry, func() error {
					updating, err := wh.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.TODO(), imageManagerWebhookCfgName, metav1.GetOptions{})
					if err != nil {
						if apierrors.IsNotFound(err) {
							klog.Info("webhook configuration not found, ", imageManagerWebhookCfgName)
							return nil
						}
						return err
					}
					updating.Webhooks = append(existing.Webhooks[:i], existing.Webhooks[i+1:]...)
					klog.Info("removing the webhook, ", imwhName)
					_, err = wh.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(context.Background(), updating, metav1.UpdateOptions{})
					if !apierrors.IsConflict(err) {
						klog.Error("Error updating MutatingWebhookConfiguration ", err)
					}
					return err
				})
			}
		}
	}
	klog.Infof("success to clean imagemanager webhook")
	return nil
}

func mutatingWebhookName() string {
	// should be a domain with at least three segments separated by dots
	return mutatingWebhookNamePrefix + "." + constants.Namespace + ".ns"
}

func imageManagerWebhookName() string {
	return imageManagerWebhookPrefix + "." + constants.Namespace + ".ns"
}

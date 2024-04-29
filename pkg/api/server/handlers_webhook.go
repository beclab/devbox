package server

import (
	"context"
	"fmt"

	"github.com/beclab/devbox/pkg/webhook"
	
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/klog/v2"
)

var (
	errNilAdmissionRequest = fmt.Errorf("nil admission request")
)

func (h *webhooks) devcontainer(ctx *fiber.Ctx) error {
	klog.Infof("Received mutating webhook request: Method=%v, URL=%v", ctx.Method(), ctx.OriginalURL())
	admissionRequestBody := ctx.BodyRaw()

	if len(admissionRequestBody) == 0 {
		klog.Error("Error reading admission request body, body is empty")
		return fiber.NewError(fiber.StatusBadRequest, "empty request admission request body")
	}

	var admissionReq, admissionResp admissionv1.AdmissionReview
	proxyUUID := uuid.New()
	if _, _, err := webhook.Deserializer.Decode(admissionRequestBody, nil, &admissionReq); err != nil {
		klog.Error("Error decoding admission request body, ", err)
		admissionResp.Response = h.webhook.AdmissionError(err)
	} else {
		admissionResp.Response = h.mutate(ctx.Context(), admissionReq.Request, proxyUUID)
	}

	admissionResp.TypeMeta = admissionReq.TypeMeta
	admissionResp.Kind = admissionReq.Kind

	requestForNamespace := "unknown"
	if admissionReq.Request != nil {
		requestForNamespace = admissionReq.Request.Namespace
	}

	klog.Infof("Done responding to admission request for pod with UUID %s in namespace %s", proxyUUID, requestForNamespace)
	return ctx.JSON(&admissionResp)
}

func (h *webhooks) mutate(ctx context.Context, req *admissionv1.AdmissionRequest, proxyUUID uuid.UUID) *admissionv1.AdmissionResponse {
	if req == nil {
		klog.Error("nil admission Request")
		return h.webhook.AdmissionError(errNilAdmissionRequest)
	}

	// Start building the response
	resp := &admissionv1.AdmissionResponse{
		Allowed: true,
		UID:     req.UID,
	}

	var (
		patchBytes []byte
		err        error
	)

	klog.Info("Creating patch for resource, ", req.Resource)
	switch req.Resource.Resource {
	case "deployments", "statefulsets", "daemonsets":
		patchBytes, err = h.webhook.MutateAppName(ctx, req)
		if err != nil {
			klog.Errorf("Failed to create patch for pod with UUID %s in namespace %s", proxyUUID, req.Namespace)
			return h.webhook.AdmissionError(err)
		}
	case "pods":
		patchBytes, err = h.webhook.MutatePodContainers(ctx, req.Namespace, req.Object.Raw, proxyUUID, BaseDir)
		if err != nil {
			klog.Errorf("Failed to create patch for pod with UUID %s in namespace %s, %s", proxyUUID, req.Namespace, err.Error())
			return h.webhook.AdmissionError(err)
		}
	}

	if len(patchBytes) > 0 {
		klog.Infof("Done creating patch admission response for pod with UUID %s in namespace %s", proxyUUID, req.Namespace)
		h.webhook.PatchAdmissionResponse(resp, patchBytes)
	}

	return resp

}

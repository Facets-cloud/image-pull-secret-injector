package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	ExcludeAnnotation = "image-pull-secret-injector.facets.cloud/exclude"
)

type podMutator struct {
	Log         logr.Logger
	Client      client.Client
	SecretNames []string
	decoder     *admission.Decoder
}

// InjectDecoder injects the decoder.
func (a *podMutator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=ignore,groups="",resources=pods,verbs=create,versions=v1,name=mpod.kb.io,admissionReviewVersions=v1,sideEffects=None

func (a *podMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	a.Log.Info("!!! WEBHOOK CALLED !!!", "name", req.Name, "namespace", req.Namespace, "operation", req.Operation)

	pod := &corev1.Pod{}
	err := a.decoder.Decode(req, pod)
	if err != nil {
		a.Log.Error(err, "failed to decode pod")
		return admission.Errored(http.StatusBadRequest, err)
	}

	name := req.Name
	if name == "" {
		name = pod.Name
	}
	if name == "" {
		name = pod.GenerateName + "[SERVER GENERATED]"
	}

	// Check if pod should be excluded from mutation
	if pod.Annotations != nil {
		if exclude, ok := pod.Annotations[ExcludeAnnotation]; ok && exclude == "true" {
			a.Log.Info("pod excluded from mutation via annotation", "namespace", req.Namespace, "name", name)
			return admission.Allowed("pod excluded from mutation via annotation")
		}
	}

	a.Log.Info("processing pod", "namespace", req.Namespace, "name", name, "configured_secrets", a.SecretNames)

	// Log original secrets
	originalSecrets := make([]string, 0)
	for _, s := range pod.Spec.ImagePullSecrets {
		originalSecrets = append(originalSecrets, s.Name)
	}
	a.Log.Info("original pod secrets", "secrets", originalSecrets)

	// Keep existing secrets and append our configured ones if they don't exist
	for _, secretName := range a.SecretNames {
		exists := false
		for _, existing := range pod.Spec.ImagePullSecrets {
			if existing.Name == secretName {
				exists = true
				break
			}
		}
		if !exists {
			a.Log.Info("adding configured secret", "secret_name", secretName)
			pod.Spec.ImagePullSecrets = append(pod.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: secretName})
		} else {
			a.Log.Info("secret already exists, skipping", "secret_name", secretName)
		}
	}

	// Log final secrets
	finalSecrets := make([]string, 0)
	for _, s := range pod.Spec.ImagePullSecrets {
		finalSecrets = append(finalSecrets, s.Name)
	}
	a.Log.Info("final pod secrets after mutation", "secrets", finalSecrets)

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		a.Log.Error(err, "failed to marshal pod")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

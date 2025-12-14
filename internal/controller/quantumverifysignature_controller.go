/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	qubeseciov1 "github.com/QubeSec/QubeSec/api/v1"
	"github.com/QubeSec/QubeSec/internal/signature"
)

// QuantumVerifySignatureReconciler reconciles a QuantumVerifySignature object
type QuantumVerifySignatureReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=qubesec.io,resources=quantumverifysignatures,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumverifysignatures/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumverifysignatures/finalizers,verbs=update
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumsignaturekeypairs,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// updateStatus refreshes the object and updates its status
func (r *QuantumVerifySignatureReconciler) updateStatus(ctx context.Context, qvs *qubeseciov1.QuantumVerifySignature) error {
	// Refresh the object to avoid conflicts
	if err := r.Get(ctx, client.ObjectKey{
		Name:      qvs.Name,
		Namespace: qvs.Namespace,
	}, qvs); err != nil {
		return err
	}
	return r.Status().Update(ctx, qvs)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *QuantumVerifySignatureReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the QuantumVerifySignature resource
	quantumVerifySignature := &qubeseciov1.QuantumVerifySignature{}
	if err := r.Get(ctx, req.NamespacedName, quantumVerifySignature); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if quantumVerifySignature.Spec.Algorithm == "" {
		quantumVerifySignature.Status.Status = "Failed"
		quantumVerifySignature.Status.Error = "spec.algorithm is required"
		_ = r.updateStatus(ctx, quantumVerifySignature)
		return ctrl.Result{}, fmt.Errorf("spec.algorithm is required")
	}

	// If already verified, no need to reconcile again
	if (quantumVerifySignature.Status.Status == "Valid" || quantumVerifySignature.Status.Status == "Invalid") &&
		quantumVerifySignature.Status.LastCheckedTime != nil {
		log.Info("Signature already verified, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	// Get the public key from the referenced QuantumSignatureKeyPair
	pkNamespace := quantumVerifySignature.Spec.PublicKeyRef.Namespace
	if pkNamespace == "" {
		pkNamespace = quantumVerifySignature.Namespace
	}

	sigKeyPair := &qubeseciov1.QuantumSignatureKeyPair{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      quantumVerifySignature.Spec.PublicKeyRef.Name,
		Namespace: pkNamespace,
	}, sigKeyPair); err != nil {
		log.Error(err, "Failed to get referenced QuantumSignatureKeyPair")
		quantumVerifySignature.Status.Status = "Failed"
		quantumVerifySignature.Status.Error = fmt.Sprintf("Failed to get referenced QuantumSignatureKeyPair: %v", err)
		_ = r.updateStatus(ctx, quantumVerifySignature)
		return ctrl.Result{}, err
	}

	// Get the secret containing the keys
	keySecretName := sigKeyPair.Spec.SecretName
	if keySecretName == "" {
		keySecretName = sigKeyPair.Name
	}

	keySecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      keySecretName,
		Namespace: pkNamespace,
	}, keySecret); err != nil {
		log.Error(err, "Failed to get key pair secret")
		quantumVerifySignature.Status.Status = "Failed"
		quantumVerifySignature.Status.Error = fmt.Sprintf("Failed to get key pair secret: %v", err)
		_ = r.updateStatus(ctx, quantumVerifySignature)
		return ctrl.Result{}, err
	}

	// Extract public key from secret
	publicKeyPEM, ok := keySecret.Data["public-key"]
	if !ok {
		log.Error(nil, "Public key not found in secret")
		quantumVerifySignature.Status.Status = "Failed"
		quantumVerifySignature.Status.Error = "Public key not found in secret"
		_ = r.updateStatus(ctx, quantumVerifySignature)
		return ctrl.Result{}, fmt.Errorf("public key not found in secret")
	}

	// Get the message from the referenced secret
	msgNamespace := quantumVerifySignature.Spec.MessageRef.Namespace
	if msgNamespace == "" {
		msgNamespace = quantumVerifySignature.Namespace
	}

	messageSecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      quantumVerifySignature.Spec.MessageRef.Name,
		Namespace: msgNamespace,
	}, messageSecret); err != nil {
		log.Error(err, "Failed to get message secret")
		quantumVerifySignature.Status.Status = "Failed"
		quantumVerifySignature.Status.Error = fmt.Sprintf("Failed to get message secret: %v", err)
		_ = r.updateStatus(ctx, quantumVerifySignature)
		return ctrl.Result{}, err
	}

	// Extract message from secret
	messageKey := quantumVerifySignature.Spec.MessageKey
	if messageKey == "" {
		messageKey = "message"
	}

	messageBytes, ok := messageSecret.Data[messageKey]
	if !ok {
		log.Error(nil, "Message not found in secret", "key", messageKey)
		quantumVerifySignature.Status.Status = "Failed"
		quantumVerifySignature.Status.Error = fmt.Sprintf("Message key '%s' not found in secret", messageKey)
		_ = r.updateStatus(ctx, quantumVerifySignature)
		return ctrl.Result{}, fmt.Errorf("message key '%s' not found in secret", messageKey)
	}

	// Get the signature from the referenced secret
	sigNamespace := quantumVerifySignature.Spec.SignatureRef.Namespace
	if sigNamespace == "" {
		sigNamespace = quantumVerifySignature.Namespace
	}

	signatureSecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      quantumVerifySignature.Spec.SignatureRef.Name,
		Namespace: sigNamespace,
	}, signatureSecret); err != nil {
		log.Error(err, "Failed to get signature secret")
		quantumVerifySignature.Status.Status = "Failed"
		quantumVerifySignature.Status.Error = fmt.Sprintf("Failed to get signature secret: %v", err)
		_ = r.updateStatus(ctx, quantumVerifySignature)
		return ctrl.Result{}, err
	}

	// Extract signature from secret
	signatureKey := quantumVerifySignature.Spec.SignatureKey
	if signatureKey == "" {
		signatureKey = "signature"
	}

	signatureBytes, ok := signatureSecret.Data[signatureKey]
	if !ok {
		log.Error(nil, "Signature not found in secret", "key", signatureKey)
		quantumVerifySignature.Status.Status = "Failed"
		quantumVerifySignature.Status.Error = fmt.Sprintf("Signature key '%s' not found in secret", signatureKey)
		_ = r.updateStatus(ctx, quantumVerifySignature)
		return ctrl.Result{}, fmt.Errorf("signature key '%s' not found in secret", signatureKey)
	}

	// Verify the signature
	valid, err := signature.VerifySignature(
		quantumVerifySignature.Spec.Algorithm,
		publicKeyPEM,
		messageBytes,
		signatureBytes,
		ctx,
	)
	if err != nil {
		log.Error(err, "Failed to verify signature")
		quantumVerifySignature.Status.Status = "Failed"
		quantumVerifySignature.Status.Error = fmt.Sprintf("Failed to verify signature: %v", err)
		_ = r.updateStatus(ctx, quantumVerifySignature)
		return ctrl.Result{}, err
	}

	// Update status
	// Get the latest version before updating status to avoid conflicts
	if err := r.Get(ctx, client.ObjectKey{
		Name:      quantumVerifySignature.Name,
		Namespace: quantumVerifySignature.Namespace,
	}, quantumVerifySignature); err != nil {
		log.Error(err, "Failed to get latest object before status update")
		return ctrl.Result{}, err
	}

	now := metav1.Now()
	quantumVerifySignature.Status.LastCheckedTime = &now
	quantumVerifySignature.Status.MessageFingerprint = signature.MessageFingerprint(messageBytes)
	quantumVerifySignature.Status.Verified = valid

	if valid {
		quantumVerifySignature.Status.Status = "Valid"
		quantumVerifySignature.Status.Error = ""
	} else {
		quantumVerifySignature.Status.Status = "Invalid"
		quantumVerifySignature.Status.Error = "signature verification failed"
	}

	if err := r.Status().Update(ctx, quantumVerifySignature); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	if valid {
		log.Info("Signature verified successfully")
	} else {
		log.Info("Signature verification failed")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QuantumVerifySignatureReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&qubeseciov1.QuantumVerifySignature{}).
		Complete(r)
}

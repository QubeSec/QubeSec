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

// QuantumSignMessageReconciler reconciles a QuantumSignMessage object
type QuantumSignMessageReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=qubesec.io,resources=quantumsignmessages,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumsignmessages/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumsignmessages/finalizers,verbs=update
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumsignaturekeypairs,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// updateStatus refreshes the object and updates its status
func (r *QuantumSignMessageReconciler) updateStatus(ctx context.Context, qsm *qubeseciov1.QuantumSignMessage) error {
	// Refresh the object to avoid conflicts
	if err := r.Get(ctx, client.ObjectKey{
		Name:      qsm.Name,
		Namespace: qsm.Namespace,
	}, qsm); err != nil {
		return err
	}
	return r.Status().Update(ctx, qsm)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *QuantumSignMessageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the QuantumSignMessage resource
	quantumSignMessage := &qubeseciov1.QuantumSignMessage{}
	if err := r.Get(ctx, req.NamespacedName, quantumSignMessage); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if quantumSignMessage.Spec.Algorithm == "" {
		quantumSignMessage.Status.Status = "Failed"
		quantumSignMessage.Status.Error = "spec.algorithm is required"
		_ = r.updateStatus(ctx, quantumSignMessage)
		return ctrl.Result{}, fmt.Errorf("spec.algorithm is required")
	}

	// If already successfully signed, no need to reconcile again
	if quantumSignMessage.Status.Status == "Success" && quantumSignMessage.Status.Signature != "" {
		log.Info("Message already signed, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	// Determine output secret name
	outputSecretName := quantumSignMessage.Spec.OutputSecretName
	if outputSecretName == "" {
		outputSecretName = fmt.Sprintf("%s-signature", quantumSignMessage.Name)
	}

	// Get the private key from the referenced QuantumSignatureKeyPair
	pkNamespace := quantumSignMessage.Spec.PrivateKeyRef.Namespace
	if pkNamespace == "" {
		pkNamespace = quantumSignMessage.Namespace
	}

	sigKeyPair := &qubeseciov1.QuantumSignatureKeyPair{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      quantumSignMessage.Spec.PrivateKeyRef.Name,
		Namespace: pkNamespace,
	}, sigKeyPair); err != nil {
		log.Error(err, "Failed to get referenced QuantumSignatureKeyPair")
		quantumSignMessage.Status.Status = "Failed"
		quantumSignMessage.Status.Error = fmt.Sprintf("Failed to get referenced QuantumSignatureKeyPair: %v", err)
		_ = r.updateStatus(ctx, quantumSignMessage)
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
		quantumSignMessage.Status.Status = "Failed"
		quantumSignMessage.Status.Error = fmt.Sprintf("Failed to get key pair secret: %v", err)
		_ = r.updateStatus(ctx, quantumSignMessage)
		return ctrl.Result{}, err
	}

	// Extract private key from secret
	privateKeyPEM, ok := keySecret.Data["private-key"]
	if !ok {
		log.Error(nil, "Private key not found in secret")
		quantumSignMessage.Status.Status = "Failed"
		quantumSignMessage.Status.Error = "Private key not found in secret"
		_ = r.updateStatus(ctx, quantumSignMessage)
		return ctrl.Result{}, fmt.Errorf("private key not found in secret")
	}

	// Get the message from the referenced secret
	msgNamespace := quantumSignMessage.Spec.MessageRef.Namespace
	if msgNamespace == "" {
		msgNamespace = quantumSignMessage.Namespace
	}

	messageSecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      quantumSignMessage.Spec.MessageRef.Name,
		Namespace: msgNamespace,
	}, messageSecret); err != nil {
		log.Error(err, "Failed to get message secret")
		quantumSignMessage.Status.Status = "Failed"
		quantumSignMessage.Status.Error = fmt.Sprintf("Failed to get message secret: %v", err)
		_ = r.updateStatus(ctx, quantumSignMessage)
		return ctrl.Result{}, err
	}

	// Extract message from secret
	messageKey := quantumSignMessage.Spec.MessageKey
	if messageKey == "" {
		messageKey = "message"
	}

	messageBytes, ok := messageSecret.Data[messageKey]
	if !ok {
		log.Error(nil, "Message not found in secret", "key", messageKey)
		quantumSignMessage.Status.Status = "Failed"
		quantumSignMessage.Status.Error = fmt.Sprintf("Message key '%s' not found in secret", messageKey)
		_ = r.updateStatus(ctx, quantumSignMessage)
		return ctrl.Result{}, fmt.Errorf("message key '%s' not found in secret", messageKey)
	}

	// Sign the message
	sig, err := signature.SignMessage(quantumSignMessage.Spec.Algorithm, privateKeyPEM, messageBytes, ctx)
	if err != nil {
		log.Error(err, "Failed to sign message")
		quantumSignMessage.Status.Status = "Failed"
		quantumSignMessage.Status.Error = fmt.Sprintf("Failed to sign message: %v", err)
		_ = r.updateStatus(ctx, quantumSignMessage)
		return ctrl.Result{}, err
	}

	// Create or update output secret with the signature
	outputSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      outputSecretName,
			Namespace: quantumSignMessage.Namespace,
		},
		Data: map[string][]byte{
			"signature": sig,
		},
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(quantumSignMessage, outputSecret, r.Scheme); err != nil {
		log.Error(err, "Failed to set owner reference")
		return ctrl.Result{}, err
	}

	// Create or update the secret
	existingSecret := &corev1.Secret{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      outputSecretName,
		Namespace: quantumSignMessage.Namespace,
	}, existingSecret)

	if err == nil {
		// Update existing secret
		existingSecret.Data = outputSecret.Data
		if err := r.Update(ctx, existingSecret); err != nil {
			log.Error(err, "Failed to update output secret")
			quantumSignMessage.Status.Status = "Failed"
			quantumSignMessage.Status.Error = fmt.Sprintf("Failed to update output secret: %v", err)
			_ = r.updateStatus(ctx, quantumSignMessage)
			return ctrl.Result{}, err
		}
	} else if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	} else {
		// Create new secret
		if err := r.Create(ctx, outputSecret); err != nil {
			log.Error(err, "Failed to create output secret")
			quantumSignMessage.Status.Status = "Failed"
			quantumSignMessage.Status.Error = fmt.Sprintf("Failed to create output secret: %v", err)
			_ = r.updateStatus(ctx, quantumSignMessage)
			return ctrl.Result{}, err
		}
	}

	// Update status - get the latest version first to avoid conflicts
	if err := r.Get(ctx, client.ObjectKey{
		Name:      quantumSignMessage.Name,
		Namespace: quantumSignMessage.Namespace,
	}, quantumSignMessage); err != nil {
		log.Error(err, "Failed to get latest object before status update")
		return ctrl.Result{}, err
	}

	// Now update the status
	now := metav1.Now()
	quantumSignMessage.Status.Status = "Success"
	quantumSignMessage.Status.Signature = signature.EncodeSignatureBase64(sig)
	quantumSignMessage.Status.MessageFingerprint = signature.MessageFingerprint(messageBytes)
	quantumSignMessage.Status.SignatureReference = &qubeseciov1.ObjectReference{
		Name:      outputSecretName,
		Namespace: quantumSignMessage.Namespace,
	}
	quantumSignMessage.Status.LastUpdateTime = &now
	quantumSignMessage.Status.Error = ""

	if err := r.Status().Update(ctx, quantumSignMessage); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully signed message", "signatureName", outputSecretName)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QuantumSignMessageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&qubeseciov1.QuantumSignMessage{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

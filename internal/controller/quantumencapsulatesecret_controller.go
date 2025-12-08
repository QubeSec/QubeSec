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
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	qubeseciov1 "github.com/QubeSec/QubeSec/api/v1"
	"github.com/QubeSec/QubeSec/internal/sharedsecret"
)

// QuantumEncapsulateSecretReconciler reconciles a QuantumEncapsulateSecret object
type QuantumEncapsulateSecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=qubesec.io,resources=quantumencapsulatesecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumencapsulatesecrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumencapsulatesecrets/finalizers,verbs=update
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumkemkeypairs,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *QuantumEncapsulateSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the QuantumEncapsulateSecret resource
	quantumEncapsulatedSecret := &qubeseciov1.QuantumEncapsulateSecret{}
	if err := r.Get(ctx, req.NamespacedName, quantumEncapsulatedSecret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if secret already exists - if so, skip reconciliation
	secretName := quantumEncapsulatedSecret.Spec.SecretName
	if secretName == "" {
		secretName = fmt.Sprintf("%s-shared-secret", quantumEncapsulatedSecret.Name)
	}

	existingSecret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: quantumEncapsulatedSecret.Namespace,
	}, existingSecret)

	if err == nil {
		// Secret already exists, check if status is already set
		if quantumEncapsulatedSecret.Status.Status == "Success" && quantumEncapsulatedSecret.Status.SharedSecretReference != nil {
			return ctrl.Result{}, nil
		}
		// Update status to reflect existing secret only if not already set
		if quantumEncapsulatedSecret.Status.Status != "Success" {
			// Get the ciphertext from the secret (binary data)
			ciphertextBinary := existingSecret.Data["ciphertext"]

			// Re-fetch the latest version to avoid optimistic locking conflicts
			if err := r.Get(ctx, req.NamespacedName, quantumEncapsulatedSecret); err != nil {
				log.Error(err, "Failed to re-fetch QuantumEncapsulateSecret before status update")
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			now := metav1.Now()
			quantumEncapsulatedSecret.Status.Status = "Success"
			quantumEncapsulatedSecret.Status.SharedSecretReference = &qubeseciov1.ObjectReference{
				Name:      secretName,
				Namespace: quantumEncapsulatedSecret.Namespace,
			}
			// Calculate fingerprint from the cached shared secret
			fingerprint := sha256.Sum256(existingSecret.Data["shared-secret"])
			quantumEncapsulatedSecret.Status.Fingerprint = hex.EncodeToString(fingerprint[:])[:10]
			quantumEncapsulatedSecret.Status.LastUpdateTime = &now
			// Hex-encode the binary ciphertext for status (so decapsulate can decode it)
			quantumEncapsulatedSecret.Status.Ciphertext = hex.EncodeToString(ciphertextBinary)
			quantumEncapsulatedSecret.Status.Error = ""

			if err := r.Status().Update(ctx, quantumEncapsulatedSecret); err != nil {
				log.Error(err, "Failed to update status for existing secret")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}

	// Get the referenced QuantumKEMKeyPair
	namespace := quantumEncapsulatedSecret.Spec.PublicKeyRef.Namespace
	if namespace == "" {
		namespace = quantumEncapsulatedSecret.Namespace
	}

	kemKeyPair := &qubeseciov1.QuantumKEMKeyPair{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      quantumEncapsulatedSecret.Spec.PublicKeyRef.Name,
		Namespace: namespace,
	}, kemKeyPair); err != nil {
		log.Error(err, "Failed to get referenced QuantumKEMKeyPair")
		quantumEncapsulatedSecret.Status.Status = "Failed"
		quantumEncapsulatedSecret.Status.Error = fmt.Sprintf("Failed to get referenced QuantumKEMKeyPair: %v", err)
		_ = r.Status().Update(ctx, quantumEncapsulatedSecret)
		return ctrl.Result{}, err
	}

	// Get the public key from the secret created by QuantumKEMKeyPair
	kemSecretName := kemKeyPair.Spec.SecretName
	if kemSecretName == "" {
		kemSecretName = quantumEncapsulatedSecret.Spec.PublicKeyRef.Name
	}
	secret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      kemSecretName,
		Namespace: namespace,
	}, secret); err != nil {
		log.Error(err, "Failed to get public key secret")
		quantumEncapsulatedSecret.Status.Status = "Failed"
		quantumEncapsulatedSecret.Status.Error = fmt.Sprintf("Failed to get public key secret: %v", err)
		_ = r.Status().Update(ctx, quantumEncapsulatedSecret)
		return ctrl.Result{}, err
	}

	// Extract public key from secret
	publicKeyPEM, ok := secret.Data["public-key"]
	if !ok {
		log.Error(nil, "Public key not found in secret")
		quantumEncapsulatedSecret.Status.Status = "Failed"
		quantumEncapsulatedSecret.Status.Error = "Public key not found in secret"
		_ = r.Status().Update(ctx, quantumEncapsulatedSecret)
		return ctrl.Result{}, fmt.Errorf("public key not found in secret")
	}

	// Derive shared secret
	ciphertext, sharedSecret, err := sharedsecret.DeriveSharedSecret(
		quantumEncapsulatedSecret.Spec.Algorithm,
		publicKeyPEM,
		ctx,
	)
	if err != nil {
		log.Error(err, "Failed to derive shared secret")
		quantumEncapsulatedSecret.Status.Status = "Failed"
		quantumEncapsulatedSecret.Status.Error = fmt.Sprintf("Failed to derive shared secret: %v", err)
		_ = r.Status().Update(ctx, quantumEncapsulatedSecret)
		return ctrl.Result{}, err
	}

	// Create secret with shared secret and ciphertext
	derivedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: quantumEncapsulatedSecret.Namespace,
		},
		Data: map[string][]byte{
			"shared-secret": sharedSecret,
			"ciphertext":    ciphertext,
		},
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(quantumEncapsulatedSecret, derivedSecret, r.Scheme); err != nil {
		log.Error(err, "Failed to set owner reference")
		return ctrl.Result{}, err
	}

	// Create secret
	if err := r.Create(ctx, derivedSecret); err != nil {
		log.Error(err, "Failed to create secret")
		quantumEncapsulatedSecret.Status.Status = "Failed"
		quantumEncapsulatedSecret.Status.Error = fmt.Sprintf("Failed to create secret: %v", err)
		_ = r.Status().Update(ctx, quantumEncapsulatedSecret)
		return ctrl.Result{}, err
	}

	// Update status with hex-encoded ciphertext for reference and fingerprint
	now := metav1.Now()
	quantumEncapsulatedSecret.Status.Status = "Success"
	quantumEncapsulatedSecret.Status.Ciphertext = hex.EncodeToString(ciphertext)
	// Calculate fingerprint from shared secret
	fingerprint := sha256.Sum256(sharedSecret)
	quantumEncapsulatedSecret.Status.Fingerprint = hex.EncodeToString(fingerprint[:])[:10]
	quantumEncapsulatedSecret.Status.SharedSecretReference = &qubeseciov1.ObjectReference{
		Name:      secretName,
		Namespace: quantumEncapsulatedSecret.Namespace,
	}
	quantumEncapsulatedSecret.Status.LastUpdateTime = &now
	quantumEncapsulatedSecret.Status.Error = ""

	if err := r.Status().Update(ctx, quantumEncapsulatedSecret); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully derived shared secret", "sharedSecretName", secretName)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QuantumEncapsulateSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&qubeseciov1.QuantumEncapsulateSecret{}).
		Owns(&corev1.Secret{}). // Watch Secret objects owned by QuantumEncapsulateSecret
		Named("quantumencapsulatesecret").
		Complete(r)
}

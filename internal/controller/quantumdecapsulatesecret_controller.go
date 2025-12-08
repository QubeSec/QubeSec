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

// QuantumDecapsulateSecretReconciler reconciles a QuantumDecapsulateSecret object
type QuantumDecapsulateSecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=qubesec.io,resources=quantumdecapsulatesecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumdecapsulatesecrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumdecapsulatesecrets/finalizers,verbs=update
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumkemkeypairs,verbs=get;list;watch
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumencapsulatesecrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *QuantumDecapsulateSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the QuantumDecapsulateSecret resource
	quantumDecapsulateSecret := &qubeseciov1.QuantumDecapsulateSecret{}
	if err := r.Get(ctx, req.NamespacedName, quantumDecapsulateSecret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if secret already exists - if so, skip reconciliation
	secretName := quantumDecapsulateSecret.Spec.SecretName
	if secretName == "" {
		secretName = fmt.Sprintf("%s-shared-secret", quantumDecapsulateSecret.Name)
	}

	existingSecret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: quantumDecapsulateSecret.Namespace,
	}, existingSecret)

	if err == nil {
		// Secret already exists, check if status is already set
		if quantumDecapsulateSecret.Status.Status == "Success" && quantumDecapsulateSecret.Status.SharedSecretReference != nil {
			return ctrl.Result{}, nil
		}
		// Update status to reflect existing secret only if not already set
		if quantumDecapsulateSecret.Status.Status != "Success" {
			// Re-fetch the latest version to avoid optimistic locking conflicts
			if err := r.Get(ctx, req.NamespacedName, quantumDecapsulateSecret); err != nil {
				log.Error(err, "Failed to re-fetch QuantumDecapsulateSecret before status update")
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}

			now := metav1.Now()
			quantumDecapsulateSecret.Status.Status = "Success"
			quantumDecapsulateSecret.Status.SharedSecretReference = &qubeseciov1.ObjectReference{
				Name:      secretName,
				Namespace: quantumDecapsulateSecret.Namespace,
			}
			// Calculate fingerprint from the cached shared secret
			fingerprint := sha256.Sum256(existingSecret.Data["shared-secret"])
			quantumDecapsulateSecret.Status.Fingerprint = hex.EncodeToString(fingerprint[:])[:10]
			quantumDecapsulateSecret.Status.LastUpdateTime = &now
			quantumDecapsulateSecret.Status.Error = ""

			if err := r.Status().Update(ctx, quantumDecapsulateSecret); err != nil {
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
	namespace := quantumDecapsulateSecret.Spec.PrivateKeyRef.Namespace
	if namespace == "" {
		namespace = quantumDecapsulateSecret.Namespace
	}

	kemKeyPair := &qubeseciov1.QuantumKEMKeyPair{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      quantumDecapsulateSecret.Spec.PrivateKeyRef.Name,
		Namespace: namespace,
	}, kemKeyPair); err != nil {
		log.Error(err, "Failed to get referenced QuantumKEMKeyPair")
		quantumDecapsulateSecret.Status.Status = "Failed"
		quantumDecapsulateSecret.Status.Error = fmt.Sprintf("Failed to get referenced QuantumKEMKeyPair: %v", err)
		_ = r.Status().Update(ctx, quantumDecapsulateSecret)
		return ctrl.Result{}, err
	}

	// Get the private key from the secret created by QuantumKEMKeyPair
	kemSecretName := kemKeyPair.Spec.SecretName
	if kemSecretName == "" {
		kemSecretName = quantumDecapsulateSecret.Spec.PrivateKeyRef.Name
	}
	secret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      kemSecretName,
		Namespace: namespace,
	}, secret); err != nil {
		log.Error(err, "Failed to get private key secret")
		quantumDecapsulateSecret.Status.Status = "Failed"
		quantumDecapsulateSecret.Status.Error = fmt.Sprintf("Failed to get private key secret: %v", err)
		_ = r.Status().Update(ctx, quantumDecapsulateSecret)
		return ctrl.Result{}, err
	}

	// Extract private key from secret
	privateKeyPEM, ok := secret.Data["private-key"]
	if !ok {
		log.Error(nil, "Private key not found in secret")
		quantumDecapsulateSecret.Status.Status = "Failed"
		quantumDecapsulateSecret.Status.Error = "Private key not found in secret"
		_ = r.Status().Update(ctx, quantumDecapsulateSecret)
		return ctrl.Result{}, fmt.Errorf("private key not found in secret")
	}

	// Resolve ciphertext from spec or referenced QuantumEncapsulateSecret status
	ciphertextHex := quantumDecapsulateSecret.Spec.Ciphertext
	if ciphertextHex == "" && quantumDecapsulateSecret.Spec.CiphertextRef != nil {
		ref := quantumDecapsulateSecret.Spec.CiphertextRef
		refNamespace := ref.Namespace
		if refNamespace == "" {
			refNamespace = quantumDecapsulateSecret.Namespace
		}

		qes := &qubeseciov1.QuantumEncapsulateSecret{}
		if err := r.Get(ctx, client.ObjectKey{Name: ref.Name, Namespace: refNamespace}, qes); err != nil {
			log.Error(err, "Failed to get referenced QuantumEncapsulateSecret for ciphertext")
			quantumDecapsulateSecret.Status.Status = "Failed"
			quantumDecapsulateSecret.Status.Error = fmt.Sprintf("Failed to get referenced QuantumEncapsulateSecret: %v", err)
			_ = r.Status().Update(ctx, quantumDecapsulateSecret)
			return ctrl.Result{}, err
		}

		ciphertextHex = qes.Status.Ciphertext
	}

	if ciphertextHex == "" {
		quantumDecapsulateSecret.Status.Status = "Failed"
		quantumDecapsulateSecret.Status.Error = "Ciphertext is required: provide spec.ciphertext or spec.ciphertextRef"
		_ = r.Status().Update(ctx, quantumDecapsulateSecret)
		return ctrl.Result{}, fmt.Errorf("ciphertext is required")
	}

	// Decode the ciphertext from hex
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		log.Error(err, "Failed to decode ciphertext")
		quantumDecapsulateSecret.Status.Status = "Failed"
		quantumDecapsulateSecret.Status.Error = fmt.Sprintf("Failed to decode ciphertext: %v", err)
		_ = r.Status().Update(ctx, quantumDecapsulateSecret)
		return ctrl.Result{}, err
	}

	// Decapsulate to recover shared secret
	sharedSecret, err := sharedsecret.DecapsulateSharedSecret(
		quantumDecapsulateSecret.Spec.Algorithm,
		privateKeyPEM,
		ciphertext,
		ctx,
	)
	if err != nil {
		log.Error(err, "Failed to decapsulate shared secret")
		quantumDecapsulateSecret.Status.Status = "Failed"
		quantumDecapsulateSecret.Status.Error = fmt.Sprintf("Failed to decapsulate shared secret: %v", err)
		_ = r.Status().Update(ctx, quantumDecapsulateSecret)
		return ctrl.Result{}, err
	}

	// Create secret with shared secret
	derivedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: quantumDecapsulateSecret.Namespace,
		},
		Data: map[string][]byte{
			"shared-secret": sharedSecret,
		},
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(quantumDecapsulateSecret, derivedSecret, r.Scheme); err != nil {
		log.Error(err, "Failed to set owner reference")
		return ctrl.Result{}, err
	}

	// Create secret
	if err := r.Create(ctx, derivedSecret); err != nil {
		log.Error(err, "Failed to create secret")
		quantumDecapsulateSecret.Status.Status = "Failed"
		quantumDecapsulateSecret.Status.Error = fmt.Sprintf("Failed to create secret: %v", err)
		_ = r.Status().Update(ctx, quantumDecapsulateSecret)
		return ctrl.Result{}, err
	}

	// Update status with fingerprint
	now := metav1.Now()
	quantumDecapsulateSecret.Status.Status = "Success"
	// Calculate fingerprint from recovered shared secret
	fingerprint := sha256.Sum256(sharedSecret)
	quantumDecapsulateSecret.Status.Fingerprint = hex.EncodeToString(fingerprint[:])[:10]
	quantumDecapsulateSecret.Status.SharedSecretReference = &qubeseciov1.ObjectReference{
		Name:      secretName,
		Namespace: quantumDecapsulateSecret.Namespace,
	}
	quantumDecapsulateSecret.Status.LastUpdateTime = &now
	quantumDecapsulateSecret.Status.Error = ""

	if err := r.Status().Update(ctx, quantumDecapsulateSecret); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully decapsulated shared secret", "secret", secretName)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QuantumDecapsulateSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&qubeseciov1.QuantumDecapsulateSecret{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

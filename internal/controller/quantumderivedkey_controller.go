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
	"github.com/QubeSec/QubeSec/internal/derivedkey"
)

// QuantumDerivedKeyReconciler reconciles a QuantumDerivedKey object
type QuantumDerivedKeyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=qubesec.io,resources=quantumderivedkeys,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumderivedkeys/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumderivedkeys/finalizers,verbs=update
// +kubebuilder:rbac:groups=qubesec.io,resources=quantumsharedsecrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *QuantumDerivedKeyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the QuantumDerivedKey resource
	quantumDerivedKey := &qubeseciov1.QuantumDerivedKey{}
	if err := r.Get(ctx, req.NamespacedName, quantumDerivedKey); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if secret already exists - if so, skip reconciliation
	secretName := quantumDerivedKey.Spec.SecretName
	if secretName == "" {
		secretName = fmt.Sprintf("%s-derived-key", quantumDerivedKey.Name)
	}

	existingSecret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: quantumDerivedKey.Namespace,
	}, existingSecret)

	if err == nil {
		// Secret already exists, check if status is already set
		if quantumDerivedKey.Status.Status == "Success" && quantumDerivedKey.Status.DerivedKeyReference != nil {
			return ctrl.Result{}, nil
		}
		// Update status to reflect existing secret only if not already set
		if quantumDerivedKey.Status.Status != "Success" {
			// Get the fingerprint from the secret
			fingerprint := string(existingSecret.Data["fingerprint"])

			now := metav1.Now()
			quantumDerivedKey.Status.Status = "Success"
			quantumDerivedKey.Status.DerivedKeyReference = &qubeseciov1.ObjectReference{
				Name:      secretName,
				Namespace: quantumDerivedKey.Namespace,
			}
			quantumDerivedKey.Status.LastUpdateTime = &now
			quantumDerivedKey.Status.KeyFingerprint = fingerprint
			quantumDerivedKey.Status.UsedSalt = quantumDerivedKey.Spec.Salt
			quantumDerivedKey.Status.UsedInfo = quantumDerivedKey.Spec.Info
			quantumDerivedKey.Status.Error = ""

			if err := r.Status().Update(ctx, quantumDerivedKey); err != nil {
				log.Error(err, "Failed to update status for existing secret")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}

	// Get the referenced shared secret (could be QuantumEncapsulateSecret or QuantumDecapsulateSecret)
	namespace := quantumDerivedKey.Spec.SharedSecretRef.Namespace
	if namespace == "" {
		namespace = quantumDerivedKey.Namespace
	}

	// Try to get QuantumEncapsulateSecret first
	sharedSecretRef := &qubeseciov1.ObjectReference{}
	sharedSecretStatus := ""

	encapsulateSecret := &qubeseciov1.QuantumEncapsulateSecret{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      quantumDerivedKey.Spec.SharedSecretRef.Name,
		Namespace: namespace,
	}, encapsulateSecret)

	if err == nil {
		// Found QuantumEncapsulateSecret
		sharedSecretRef = encapsulateSecret.Status.SharedSecretReference
		sharedSecretStatus = encapsulateSecret.Status.Status
	} else {
		// Try to get QuantumDecapsulateSecret
		decapsulateSecret := &qubeseciov1.QuantumDecapsulateSecret{}
		if err := r.Get(ctx, client.ObjectKey{
			Name:      quantumDerivedKey.Spec.SharedSecretRef.Name,
			Namespace: namespace,
		}, decapsulateSecret); err != nil {
			log.Error(err, "Failed to get referenced QuantumEncapsulateSecret or QuantumDecapsulateSecret")
			quantumDerivedKey.Status.Status = "Failed"
			quantumDerivedKey.Status.Error = fmt.Sprintf("Failed to get referenced shared secret: %v", err)
			_ = r.Status().Update(ctx, quantumDerivedKey)
			return ctrl.Result{}, err
		}
		sharedSecretRef = decapsulateSecret.Status.SharedSecretReference
		sharedSecretStatus = decapsulateSecret.Status.Status
	}

	// Check if shared secret is ready
	if sharedSecretStatus != "Success" {
		log.Info("Shared secret not ready yet, waiting...")
		quantumDerivedKey.Status.Status = "Pending"
		quantumDerivedKey.Status.Error = "Shared secret not ready"
		_ = r.Status().Update(ctx, quantumDerivedKey)
		return ctrl.Result{}, fmt.Errorf("shared secret not ready")
	}

	// Get the shared secret from the referenced secret
	secretRef := sharedSecretRef
	if secretRef == nil {
		log.Error(nil, "Shared secret reference not set")
		quantumDerivedKey.Status.Status = "Failed"
		quantumDerivedKey.Status.Error = "Shared secret reference not set"
		_ = r.Status().Update(ctx, quantumDerivedKey)
		return ctrl.Result{}, fmt.Errorf("shared secret reference not set")
	}

	secret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      secretRef.Name,
		Namespace: secretRef.Namespace,
	}, secret); err != nil {
		log.Error(err, "Failed to get shared secret data")
		quantumDerivedKey.Status.Status = "Failed"
		quantumDerivedKey.Status.Error = fmt.Sprintf("Failed to get shared secret data: %v", err)
		_ = r.Status().Update(ctx, quantumDerivedKey)
		return ctrl.Result{}, err
	}

	// Extract shared secret from secret
	sharedSecretBytes, ok := secret.Data["shared-secret"]
	if !ok {
		log.Error(nil, "Shared secret not found in secret")
		quantumDerivedKey.Status.Status = "Failed"
		quantumDerivedKey.Status.Error = "Shared secret not found in secret"
		_ = r.Status().Update(ctx, quantumDerivedKey)
		return ctrl.Result{}, fmt.Errorf("shared secret not found in secret")
	}

	// Decode salt and info from hex
	salt := []byte{}
	if quantumDerivedKey.Spec.Salt != "" {
		var err error
		salt, err = hex.DecodeString(quantumDerivedKey.Spec.Salt)
		if err != nil {
			log.Error(err, "Failed to decode salt")
			quantumDerivedKey.Status.Status = "Failed"
			quantumDerivedKey.Status.Error = fmt.Sprintf("Failed to decode salt: %v", err)
			_ = r.Status().Update(ctx, quantumDerivedKey)
			return ctrl.Result{}, err
		}
	}

	info := []byte{}
	if quantumDerivedKey.Spec.Info != "" {
		var err error
		info, err = hex.DecodeString(quantumDerivedKey.Spec.Info)
		if err != nil {
			log.Error(err, "Failed to decode info")
			quantumDerivedKey.Status.Status = "Failed"
			quantumDerivedKey.Status.Error = fmt.Sprintf("Failed to decode info: %v", err)
			_ = r.Status().Update(ctx, quantumDerivedKey)
			return ctrl.Result{}, err
		}
	}

	// Derive AES-256 key
	derivedKey, err := derivedkey.DeriveAES256Key(sharedSecretBytes, salt, info, ctx)
	if err != nil {
		log.Error(err, "Failed to derive key")
		quantumDerivedKey.Status.Status = "Failed"
		quantumDerivedKey.Status.Error = fmt.Sprintf("Failed to derive key: %v", err)
		_ = r.Status().Update(ctx, quantumDerivedKey)
		return ctrl.Result{}, err
	}

	// Calculate fingerprint of derived key
	hash := sha256.Sum256(derivedKey)
	fingerprint := hex.EncodeToString(hash[:])

	// Create secret with derived key
	derivedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: quantumDerivedKey.Namespace,
		},
		Data: map[string][]byte{
			"derived-key": derivedKey,
			"fingerprint": []byte(fingerprint),
			"key-type":    []byte(quantumDerivedKey.Spec.KeyType),
		},
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(quantumDerivedKey, derivedSecret, r.Scheme); err != nil {
		log.Error(err, "Failed to set owner reference")
		return ctrl.Result{}, err
	}

	// Create secret
	if err := r.Create(ctx, derivedSecret); err != nil {
		log.Error(err, "Failed to create secret")
		quantumDerivedKey.Status.Status = "Failed"
		quantumDerivedKey.Status.Error = fmt.Sprintf("Failed to create secret: %v", err)
		_ = r.Status().Update(ctx, quantumDerivedKey)
		return ctrl.Result{}, err
	}

	// Update status
	now := metav1.Now()
	quantumDerivedKey.Status.Status = "Success"
	quantumDerivedKey.Status.DerivedKeyReference = &qubeseciov1.ObjectReference{
		Name:      secretName,
		Namespace: quantumDerivedKey.Namespace,
	}
	quantumDerivedKey.Status.LastUpdateTime = &now
	quantumDerivedKey.Status.KeyFingerprint = fingerprint
	// Set fingerprint preview (first 10 characters) for quick verification
	if len(fingerprint) >= 10 {
		quantumDerivedKey.Status.Fingerprint = fingerprint[:10]
	} else {
		quantumDerivedKey.Status.Fingerprint = fingerprint
	}
	quantumDerivedKey.Status.UsedSalt = quantumDerivedKey.Spec.Salt
	quantumDerivedKey.Status.UsedInfo = quantumDerivedKey.Spec.Info
	quantumDerivedKey.Status.Error = ""

	if err := r.Status().Update(ctx, quantumDerivedKey); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully derived key", "keyName", secretName, "keyType", quantumDerivedKey.Spec.KeyType)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QuantumDerivedKeyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&qubeseciov1.QuantumDerivedKey{}).
		Owns(&corev1.Secret{}). // Watch Secret objects owned by QuantumDerivedKey
		Named("quantumderivedkey").
		Complete(r)
}

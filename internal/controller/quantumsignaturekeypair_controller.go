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
	"sigs.k8s.io/controller-runtime/pkg/log"

	qubeseciov1 "github.com/QubeSec/QubeSec/api/v1"
	"github.com/QubeSec/QubeSec/internal/keypair"
)

// QuantumSignatureKeyPairReconciler reconciles a QuantumSignatureKeyPair object
type QuantumSignatureKeyPairReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=qubesec.io,resources=quantumsignaturekeypairs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=qubesec.io,resources=quantumsignaturekeypairs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=qubesec.io,resources=quantumsignaturekeypairs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the QuantumSignatureKeyPair object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *QuantumSignatureKeyPairReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Create QuantumSignatureKeyPair object
	quantumSignatureKeyPair := &qubeseciov1.QuantumSignatureKeyPair{}

	// Get QuantumSignatureKeyPair object
	err := r.Get(ctx, req.NamespacedName, quantumSignatureKeyPair)
	if err != nil {
		log.Error(err, "Failed to get QuantumSignatureKeyPair")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create or Update Secret
	err = r.CreateOrUpdateSecret(quantumSignatureKeyPair, ctx)
	if err != nil {
		log.Error(err, "Failed to Create or Update Secret")
		quantumSignatureKeyPair.Status.Status = "Failed"
		quantumSignatureKeyPair.Status.Error = err.Error()
		_ = r.Status().Update(ctx, quantumSignatureKeyPair)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QuantumSignatureKeyPairReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&qubeseciov1.QuantumSignatureKeyPair{}).
		Owns(&corev1.Secret{}). // Watch Secret objects owned by QuantumKEMKeyPair
		Complete(r)
}

func (r *QuantumSignatureKeyPairReconciler) CreateOrUpdateSecret(quantumSignatureKeyPair *qubeseciov1.QuantumSignatureKeyPair, ctx context.Context) error {
	// Setup logger
	log := log.FromContext(ctx)

	if quantumSignatureKeyPair.Spec.Algorithm == "" {
		return fmt.Errorf("spec.algorithm is required")
	}

	secretName := quantumSignatureKeyPair.Spec.SecretName
	if secretName == "" {
		secretName = quantumSignatureKeyPair.Name
	}

	// Create Secret object
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: quantumSignatureKeyPair.Namespace,
		},
	}

	// Get Secret object
	err := r.Get(ctx, client.ObjectKey{Namespace: secret.Namespace, Name: secret.Name}, secret)
	if err != nil && client.IgnoreNotFound(err) != nil {
		return err
	}

	// If Secret already exists, verify contents and update status
	if err == nil {
		publicKey, hasPub := secret.Data["public-key"]
		_, hasPriv := secret.Data["private-key"]
		if !hasPub || !hasPriv {
			quantumSignatureKeyPair.Status.Status = "Failed"
			quantumSignatureKeyPair.Status.Error = fmt.Sprintf("existing secret %s missing public-key/private-key", secretName)
			_ = r.Status().Update(ctx, quantumSignatureKeyPair)
			return fmt.Errorf("existing secret %s missing public-key/private-key", secretName)
		}

		fingerprint := sha256.Sum256(publicKey)

		if quantumSignatureKeyPair.Status.Status != "Success" || quantumSignatureKeyPair.Status.PublicKeyFingerprint == "" {
			if err := r.Get(ctx, client.ObjectKey{Namespace: quantumSignatureKeyPair.Namespace, Name: quantumSignatureKeyPair.Name}, quantumSignatureKeyPair); err != nil {
				log.Error(err, "Failed to re-fetch QuantumSignatureKeyPair before status update")
				return client.IgnoreNotFound(err)
			}

			now := metav1.Now()
			quantumSignatureKeyPair.Status.Status = "Success"
			quantumSignatureKeyPair.Status.KeyPairReference = &qubeseciov1.ObjectReference{
				Name:      secretName,
				Namespace: quantumSignatureKeyPair.Namespace,
			}
			quantumSignatureKeyPair.Status.LastUpdateTime = &now
			quantumSignatureKeyPair.Status.PublicKeyFingerprint = hex.EncodeToString(fingerprint[:])[:10]
			quantumSignatureKeyPair.Status.Error = ""
			if statusErr := r.Status().Update(ctx, quantumSignatureKeyPair); statusErr != nil {
				log.Error(statusErr, "Failed to update status for existing secret")
				return statusErr
			}
		}
		return nil
	}

	// If Secret doesn't exist, create it
	publicKey, privateKey, genErr := keypair.GenerateSIGKeyPair(quantumSignatureKeyPair.Spec.Algorithm, ctx)
	if genErr != nil {
		log.Error(genErr, "Failed to generate signature keypair")
		quantumSignatureKeyPair.Status.Status = "Failed"
		quantumSignatureKeyPair.Status.Error = genErr.Error()
		_ = r.Status().Update(ctx, quantumSignatureKeyPair)
		return genErr
	}

	// Create Secret object
	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: quantumSignatureKeyPair.Namespace,
		},
		Data: map[string][]byte{
			"public-key":  []byte(publicKey),
			"private-key": []byte(privateKey),
		},
	}

	// Set owner reference to QuantumSignatureKeyPair for Secret
	err = ctrl.SetControllerReference(quantumSignatureKeyPair, newSecret, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to Set Controller Reference")
		return err
	}

	// Create Secret
	err = r.Create(ctx, newSecret)
	if err != nil {
		log.Error(err, "Failed to Create Secret")
		quantumSignatureKeyPair.Status.Status = "Failed"
		quantumSignatureKeyPair.Status.Error = err.Error()
		_ = r.Status().Update(ctx, quantumSignatureKeyPair)
		return err
	}
	log.Info("Created Secret")

	fingerprint := sha256.Sum256([]byte(publicKey))

	// Update status to Success
	now := metav1.Now()
	quantumSignatureKeyPair.Status.Status = "Success"
	quantumSignatureKeyPair.Status.KeyPairReference = &qubeseciov1.ObjectReference{
		Name:      secretName,
		Namespace: quantumSignatureKeyPair.Namespace,
	}
	quantumSignatureKeyPair.Status.PublicKeyFingerprint = hex.EncodeToString(fingerprint[:])[:10]
	quantumSignatureKeyPair.Status.LastUpdateTime = &now
	quantumSignatureKeyPair.Status.Error = ""
	if statusErr := r.Status().Update(ctx, quantumSignatureKeyPair); statusErr != nil {
		log.Error(statusErr, "Failed to update status")
		return statusErr
	}

	return nil
}

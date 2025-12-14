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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	qubeseciov1 "github.com/QubeSec/QubeSec/api/v1"
	"github.com/QubeSec/QubeSec/internal/keypair"
)

// QuantumKEMKeyPairReconciler reconciles a QuantumKEMKeyPair object
type QuantumKEMKeyPairReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=qubesec.io,resources=quantumkemkeypairs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=qubesec.io,resources=quantumkemkeypairs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=qubesec.io,resources=quantumkemkeypairs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the QuantumKEMKeyPair object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *QuantumKEMKeyPairReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Create QuantumKEMKeyPair object
	quantumKEMKeyPair := &qubeseciov1.QuantumKEMKeyPair{}

	// Get QuantumKEMKeyPair object
	err := r.Get(ctx, req.NamespacedName, quantumKEMKeyPair)
	if err != nil {
		log.Error(err, "Failed to get QuantumKEMKeyPair")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create or Update Secret
	err = r.CreateOrUpdateSecret(quantumKEMKeyPair, ctx)
	if err != nil {
		log.Error(err, "Failed to Create or Update Secret")
		quantumKEMKeyPair.Status.Status = "Failed"
		quantumKEMKeyPair.Status.Error = err.Error()
		_ = r.Status().Update(ctx, quantumKEMKeyPair)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QuantumKEMKeyPairReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&qubeseciov1.QuantumKEMKeyPair{}).
		Owns(&corev1.Secret{}). // Watch Secret objects owned by QuantumKEMKeyPair
		Complete(r)
}

func (r *QuantumKEMKeyPairReconciler) CreateOrUpdateSecret(quantumKEMKeyPair *qubeseciov1.QuantumKEMKeyPair, ctx context.Context) error {
	// Setup logger
	log := log.FromContext(ctx)

	secretName := quantumKEMKeyPair.Spec.SecretName
	if secretName == "" {
		secretName = quantumKEMKeyPair.Name
	}

	// Create Secret object
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: quantumKEMKeyPair.Namespace,
		},
	}

	// Get Secret object
	err := r.Get(ctx, client.ObjectKey{Namespace: secret.Namespace, Name: secret.Name}, secret)
	if err != nil && client.IgnoreNotFound(err) != nil {
		return err
	}

	// If Secret already exists, update status to Success
	if err == nil {
		if quantumKEMKeyPair.Status.Status != "Success" {
			now := metav1.Now()
			quantumKEMKeyPair.Status.Status = "Success"
			quantumKEMKeyPair.Status.KeyPairReference = &qubeseciov1.ObjectReference{
				Name:      secretName,
				Namespace: quantumKEMKeyPair.Namespace,
			}
			quantumKEMKeyPair.Status.LastUpdateTime = &now
			quantumKEMKeyPair.Status.Error = ""
			_ = r.Status().Update(ctx, quantumKEMKeyPair)
		}
		return nil
	}

	// If Secret doesn't exist, create it
	publicKey, privateKey, genErr := keypair.GenerateKEMKeyPair(quantumKEMKeyPair.Spec.Algorithm, ctx)
	if genErr != nil {
		log.Error(genErr, "Failed to generate KEM keypair")
		quantumKEMKeyPair.Status.Status = "Failed"
		quantumKEMKeyPair.Status.Error = genErr.Error()
		_ = r.Status().Update(ctx, quantumKEMKeyPair)
		return genErr
	}

	// Create Secret object
	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: quantumKEMKeyPair.Namespace,
		},
		Data: map[string][]byte{
			"public-key":  []byte(publicKey),
			"private-key": []byte(privateKey),
		},
	}

	// Set owner reference to QuantumKEMKeyPair for Secret
	err = ctrl.SetControllerReference(quantumKEMKeyPair, newSecret, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to Set Controller Reference")
		return err
	}

	// Create Secret
	err = r.Create(ctx, newSecret)
	if err != nil {
		log.Error(err, "Failed to Create Secret")
		return err
	}
	log.Info("Created Secret")

	// Update status to Success
	now := metav1.Now()
	quantumKEMKeyPair.Status.Status = "Success"
	quantumKEMKeyPair.Status.KeyPairReference = &qubeseciov1.ObjectReference{
		Name:      secretName,
		Namespace: quantumKEMKeyPair.Namespace,
	}
	quantumKEMKeyPair.Status.LastUpdateTime = &now
	quantumKEMKeyPair.Status.Error = ""
	_ = r.Status().Update(ctx, quantumKEMKeyPair)

	return nil
}

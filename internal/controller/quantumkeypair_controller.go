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
	"bytes"
	"context"
	"encoding/pem"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	qubeseciov1 "github.com/QubeSec/QubeSec/api/v1"
	"github.com/open-quantum-safe/liboqs-go/oqs"
)

// QuantumKeyPairReconciler reconciles a QuantumKeyPair object
type QuantumKeyPairReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=qubesec.io,resources=quantumkeypairs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=qubesec.io,resources=quantumkeypairs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=qubesec.io,resources=quantumkeypairs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the QuantumKeyPair object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *QuantumKeyPairReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Create QuantumKeyPair object
	quantumKeyPair := &qubeseciov1.QuantumKeyPair{}

	// Get QuantumKeyPair object
	err := r.Get(ctx, req.NamespacedName, quantumKeyPair)
	if err != nil {
		log.Error(err, "Failed to get QuantumKeyPair")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	quantumKeys := oqs.KeyEncapsulation{}
	defer quantumKeys.Clean() // clean up even in case of panic

	quantumAlgorithm := quantumKeyPair.Spec.Algorithm

	// Initialize liboqs-go
	if err := quantumKeys.Init(quantumAlgorithm, nil); err != nil {
		log.Error(err, "Failed to initialize liboqs-go")
		return ctrl.Result{}, err
	}

	// Generate key pair
	quantumPublicKey, err := quantumKeys.GenerateKeyPair()
	if err != nil {
		log.Error(err, "Failed to generate key pair")
		return ctrl.Result{}, err
	}

	// Generate PEM block
	publicKeyBlock := &pem.Block{
		Type:  quantumAlgorithm + " PUBLIC KEY",
		Bytes: quantumPublicKey,
	}

	// Encode public key
	var publicKeyRow bytes.Buffer
	err = pem.Encode(&publicKeyRow, publicKeyBlock)
	if err != nil {
		log.Error(err, "Failed to encode public key")
		return ctrl.Result{}, err
	}

	// Export private key
	quantumPrivateKey := quantumKeys.ExportSecretKey()

	// Generate PEM block
	privateKeyBlock := &pem.Block{
		Type:  quantumAlgorithm + " SECRET KEY",
		Bytes: quantumPrivateKey,
	}

	// Encode private key
	var privateKeyRow bytes.Buffer
	err = pem.Encode(&privateKeyRow, privateKeyBlock)
	if err != nil {
		log.Error(err, "Failed to encode private key")
		return ctrl.Result{}, err
	}

	// Create Secret object
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      quantumKeyPair.Name,
			Namespace: quantumKeyPair.Namespace,
		},
		StringData: map[string]string{
			"quantumPublicKey":  publicKeyRow.String(),
			"quantumPrivateKey": privateKeyRow.String(),
		},
	}

	// Set owner reference to QuantumKeyPair for Secret
	if err := ctrl.SetControllerReference(quantumKeyPair, secret, r.Scheme); err != nil {
		log.Error(err, "Failed to Set Controller Reference")
		return ctrl.Result{}, err
	}

	// Create Secret
	err = r.Create(ctx, secret)
	if err != nil {
		log.Error(err, "Failed to Create Secret")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QuantumKeyPairReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&qubeseciov1.QuantumKeyPair{}).
		Owns(&corev1.Secret{}). // Watch Secret objects owned by QuantumKeyPair
		Complete(r)
}

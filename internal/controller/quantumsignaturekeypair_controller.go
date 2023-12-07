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

// QuantumSignatureKeyPairReconciler reconciles a QuantumSignatureKeyPair object
type QuantumSignatureKeyPairReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=qubesec.io,resources=quantumsignaturekeypairs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=qubesec.io,resources=quantumsignaturekeypairs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=qubesec.io,resources=quantumsignaturekeypairs/finalizers,verbs=update

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

	quantumSignatureKeys := oqs.Signature{}
	defer quantumSignatureKeys.Clean() // clean up even in case of panic

	quantumSignatureAlgorithm := quantumSignatureKeyPair.Spec.Algorithm

	// Initialize liboqs-go
	if err := quantumSignatureKeys.Init(quantumSignatureAlgorithm, nil); err != nil {
		log.Error(err, "Failed to initialize liboqs-go")
		return ctrl.Result{}, err
	}

	// Generate key pair
	quantumSignaturePublicKey, err := quantumSignatureKeys.GenerateKeyPair()
	if err != nil {
		log.Error(err, "Failed to generate key pair")
		return ctrl.Result{}, err
	}

	// Generate PEM block
	publicKeyBlock := &pem.Block{
		Type:  quantumSignatureAlgorithm + " PUBLIC KEY",
		Bytes: quantumSignaturePublicKey,
	}

	// Encode public key
	var publicKeyRow bytes.Buffer
	err = pem.Encode(&publicKeyRow, publicKeyBlock)
	if err != nil {
		log.Error(err, "Failed to encode public key")
		return ctrl.Result{}, err
	}

	// Export private key
	quantumPrivateKey := quantumSignatureKeys.ExportSecretKey()

	// Generate PEM block
	privateKeyBlock := &pem.Block{
		Type:  quantumSignatureAlgorithm + " SECRET KEY",
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
			Name:      quantumSignatureKeyPair.Name,
			Namespace: quantumSignatureKeyPair.Namespace,
		},
		StringData: map[string]string{
			"quantumPublicKey":  publicKeyRow.String(),
			"quantumPrivateKey": privateKeyRow.String(),
		},
	}

	// Set owner reference to QuantumKeyPair for Secret
	if err := ctrl.SetControllerReference(quantumSignatureKeyPair, secret, r.Scheme); err != nil {
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
func (r *QuantumSignatureKeyPairReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&qubeseciov1.QuantumSignatureKeyPair{}).
		Owns(&corev1.Secret{}). // Watch Secret objects owned by QuantumKeyPair
		Complete(r)
}

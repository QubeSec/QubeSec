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
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	qubeseciov1 "github.com/QubeSec/QubeSec/api/v1"
	oqsrand "github.com/open-quantum-safe/liboqs-go/oqs/rand"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// QuantumRandomNumberReconciler reconciles a QuantumRandomNumber object
type QuantumRandomNumberReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=qubesec.io,resources=quantumrandomnumbers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=qubesec.io,resources=quantumrandomnumbers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=qubesec.io,resources=quantumrandomnumbers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the QuantumRandomNumber object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *QuantumRandomNumberReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Setup logger
	log := log.FromContext(ctx)

	// Create QuantumRandomNumber object
	quantumrandomnumber := &qubeseciov1.QuantumRandomNumber{}

	// Get QuantumRandomNumber object
	err := r.Get(ctx, req.NamespacedName, quantumrandomnumber)
	if err != nil {
		log.Error(err, "Failed to get QuantumRandomNumber")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create or Update Secret object
	if err = r.CreateOrUpdateSecret(quantumrandomnumber, ctx); err != nil {
		log.Error(err, "Failed to Create or Update Secret")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QuantumRandomNumberReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&qubeseciov1.QuantumRandomNumber{}).
		Owns(&corev1.Secret{}). // Watch Secret objects owned by QuantumRandomNumber
		Complete(r)
}

// CreateOrUpdateSecret creates or updates a Secret object with a quantum random number
func (r *QuantumRandomNumberReconciler) CreateOrUpdateSecret(quantumrandomnumber *qubeseciov1.QuantumRandomNumber, ctx context.Context) error {
	// Setup logger
	log := log.FromContext(ctx)

	// Create Secret object
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      quantumrandomnumber.Name,
			Namespace: quantumrandomnumber.Namespace,
		},
	}

	// Get Secret object
	err := r.Get(ctx, client.ObjectKey{Namespace: secret.Namespace, Name: secret.Name}, secret)
	if err != nil && client.IgnoreNotFound(err) != nil {
		return err
	}

	// If Secret doesn't exist, create it
	if err != nil {

		secret := r.GenerateRandomNumberSecret(quantumrandomnumber, ctx)

		// Create Secret
		err = r.Create(ctx, &secret)
		if err != nil {
			return err
		}
		log.Info("Created Secret")

		if err := r.UpdateStatus(quantumrandomnumber, ctx); err != nil {
			log.Error(err, "Create: Failed to Update Status")
			return err
		}

	} else {

		// If Secret exists, compair the desired state of bytes
		if quantumrandomnumber.Status.Bytes != quantumrandomnumber.Spec.Bytes ||
			quantumrandomnumber.Status.Algorithm != quantumrandomnumber.Spec.Algorithm {

			secret := r.GenerateRandomNumberSecret(quantumrandomnumber, ctx)

			// Update Secret
			err = r.Update(ctx, &secret)
			if err != nil {
				return err
			}
			log.Info("Updated Secret")

			if err := r.UpdateStatus(quantumrandomnumber, ctx); err != nil {
				log.Error(err, "Update: Failed to Update Status")
				return err
			}

		}
	}

	return nil
}

// generate random number secret
func (r *QuantumRandomNumberReconciler) GenerateRandomNumberSecret(quantumrandomnumber *qubeseciov1.QuantumRandomNumber, ctx context.Context) corev1.Secret {
	// Setup logger
	log := log.FromContext(ctx)

	// if Bytes is not set, set it to 32
	if quantumrandomnumber.Spec.Bytes == 0 {
		quantumrandomnumber.Spec.Bytes = 32
		r.Update(ctx, quantumrandomnumber)
	}

	if quantumrandomnumber.Spec.Algorithm == "" {
		quantumrandomnumber.Spec.Algorithm = "NIST-KAT"
		r.Update(ctx, quantumrandomnumber)
	}

	// Set algorithm for quantum random number
	oqsrand.RandomBytesSwitchAlgorithm(quantumrandomnumber.Spec.Algorithm)

	// Generate quantum random number
	randomNumber := oqsrand.RandomBytes(quantumrandomnumber.Spec.Bytes)

	// delete this in future
	fmt.Println(randomNumber)

	// Create Secret object with quantum random number
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      quantumrandomnumber.Name,
			Namespace: quantumrandomnumber.Namespace,
		},
		StringData: map[string]string{
			"quantumrandomnumber": string(randomNumber),
		},
	}

	// Set owner reference to QuantumRandomNumber for Secret
	if err := ctrl.SetControllerReference(quantumrandomnumber, secret, r.Scheme); err != nil {
		log.Error(err, "Failed to Set Controller Reference")
	}

	return *secret
}

// Update Status of QuantumRandomNumber
func (r *QuantumRandomNumberReconciler) UpdateStatus(quantumrandomnumber *qubeseciov1.QuantumRandomNumber, ctx context.Context) error {
	// Setup logger
	log := log.FromContext(ctx)

	// Update status of quantumrandomnumber to reflect the number of bytes of key material generated
	quantumrandomnumber.Status.Bytes = quantumrandomnumber.Spec.Bytes
	quantumrandomnumber.Status.Algorithm = quantumrandomnumber.Spec.Algorithm
	err := r.Status().Update(ctx, quantumrandomnumber)
	if err != nil {
		return err
	}
	log.Info("Updated Quantum Random Number Status")

	return nil
}

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
	"github.com/QubeSec/QubeSec/internal/certificate"
)

// QuantumCertificateReconciler reconciles a QuantumCertificate object
type QuantumCertificateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=qubesec.io,resources=quantumcertificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=qubesec.io,resources=quantumcertificates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=qubesec.io,resources=quantumcertificates/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the QuantumCertificate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *QuantumCertificateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Create QuantumCertificate object
	quantumCertificate := &qubeseciov1.QuantumCertificate{}

	// Get QuantumCertificate object
	err := r.Get(ctx, req.NamespacedName, quantumCertificate)
	if err != nil {
		log.Error(err, "Failed to get QuantumCertificate")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create or Update Secret
	err = r.CreateOrUpdateSecret(quantumCertificate, ctx)
	if err != nil {
		log.Error(err, "Failed to Create or Update Secret")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QuantumCertificateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&qubeseciov1.QuantumCertificate{}).
		Owns(&corev1.Secret{}). // Watch Secret objects owned by QuantumCertificate
		Complete(r)
}

func (r *QuantumCertificateReconciler) CreateOrUpdateSecret(QuantumCertificate *qubeseciov1.QuantumCertificate, ctx context.Context) error {
	// Setup logger
	log := log.FromContext(ctx)

	// Create Secret object
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      QuantumCertificate.Name,
			Namespace: QuantumCertificate.Namespace,
		},
	}

	// Get Secret object
	err := r.Get(ctx, client.ObjectKey{Namespace: secret.Namespace, Name: secret.Name}, secret)
	if err != nil && client.IgnoreNotFound(err) != nil {
		return err
	}

	// If Secret doesn't exist, create it
	if err != nil {

		// Generate key pair
		publicKey, privateKey := certificate.Certificate(
			QuantumCertificate.Spec.Algorithm,
			QuantumCertificate.Spec.Domain,
			QuantumCertificate.Spec.Days,
		)

		// Create Secret object
		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      QuantumCertificate.Name,
				Namespace: QuantumCertificate.Namespace,
			},
			StringData: map[string]string{
				"tls.crt": publicKey,
				"tls.key": privateKey,
			},
		}

		// Set owner reference to QuantumCertificate for Secret
		err := ctrl.SetControllerReference(QuantumCertificate, newSecret, r.Scheme)
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
	}

	return nil
}
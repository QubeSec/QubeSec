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
	log := log.FromContext(ctx)

	quantumrandomnumber := &qubeseciov1.QuantumRandomNumber{}
	err := r.Get(ctx, req.NamespacedName, quantumrandomnumber)
	if err != nil {
		log.Error(err, "Failed to get QuantumRandomNumber")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

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
		Owns(&corev1.Secret{}).
		Complete(r)
}

func (r *QuantumRandomNumberReconciler) CreateOrUpdateSecret(quantumrandomnumber *qubeseciov1.QuantumRandomNumber, ctx context.Context) error {
	log := log.FromContext(ctx)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      quantumrandomnumber.Name,
			Namespace: quantumrandomnumber.Namespace,
		},
	}

	err := r.Get(ctx, client.ObjectKey{Namespace: secret.Namespace, Name: secret.Name}, secret)
	if err != nil && client.IgnoreNotFound(err) != nil {
		return err
	}

	// If Secret doesn't exist, create it
	if err != nil {

		randomNumber := oqsrand.RandomBytes(quantumrandomnumber.Spec.Bytes)

		// delete this in future
		fmt.Println(randomNumber)

		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      quantumrandomnumber.Name,
				Namespace: quantumrandomnumber.Namespace,
			},
			StringData: map[string]string{
				"quantumrandomnumber": string(randomNumber),
			},
		}

		if err := ctrl.SetControllerReference(quantumrandomnumber, newSecret, r.Scheme); err != nil {
			log.Error(err, "Failed to Set Controller Reference")
			return err
		}

		err = r.Create(ctx, newSecret)
		if err != nil {
			return err
		}
		log.Info("Created Secret")

		// Update status of quantumrandomnumber to reflect the number of bytes of key material generated
		quantumrandomnumber.Status.Bytes = quantumrandomnumber.Spec.Bytes
		err = r.Status().Update(ctx, quantumrandomnumber)
		if err != nil {
			return err
		}
		log.Info("Updated Quantum Random Number Status")
	} else {
		// If Secret exists, update it if the number of bytes of key material generated is different
		if quantumrandomnumber.Status.Bytes != quantumrandomnumber.Spec.Bytes {

			randomNumber := oqsrand.RandomBytes(quantumrandomnumber.Spec.Bytes)

			// delete this in future
			fmt.Println(randomNumber)

			updatedSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      quantumrandomnumber.Name,
					Namespace: quantumrandomnumber.Namespace,
				},
				StringData: map[string]string{
					"quantumrandomnumber": string(randomNumber),
				},
			}

			if err := ctrl.SetControllerReference(quantumrandomnumber, updatedSecret, r.Scheme); err != nil {
				log.Error(err, "Failed to Set Controller Reference")
				return err
			}

			err = r.Update(ctx, updatedSecret)
			if err != nil {
				return err
			}
			log.Info("Updated Secret")

			quantumrandomnumber.Status.Bytes = quantumrandomnumber.Spec.Bytes
			err = r.Status().Update(ctx, quantumrandomnumber)
			if err != nil {
				return err
			}
			log.Info("Updated Quantum Random Number Status")
		}
	}

	return nil
}

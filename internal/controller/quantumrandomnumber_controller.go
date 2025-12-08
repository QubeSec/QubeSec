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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	qubeseciov1 "github.com/QubeSec/QubeSec/api/v1"
	"github.com/open-quantum-safe/liboqs-go/oqs"

	"github.com/QubeSec/QubeSec/internal/shannonentropy"
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
	quantumRandomNumber := &qubeseciov1.QuantumRandomNumber{}

	// Get QuantumRandomNumber object
	err := r.Get(ctx, req.NamespacedName, quantumRandomNumber)
	if err != nil {
		log.Error(err, "Failed to get QuantumRandomNumber")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Apply defaults if not already set
	r.applyDefaults(quantumRandomNumber)

	// Validate the resource
	if err := r.validateQuantumRandomNumber(quantumRandomNumber); err != nil {
		log.Error(err, "QuantumRandomNumber validation failed")
		quantumRandomNumber.Status.Status = "Failed"
		quantumRandomNumber.Status.Error = err.Error()
		_ = r.Status().Update(ctx, quantumRandomNumber)
		return ctrl.Result{}, nil
	}

	// Create or Update Secret object
	err = r.CreateOrUpdateSecret(quantumRandomNumber, ctx)
	if err != nil {
		log.Error(err, "Failed to Create or Update Secret")
		quantumRandomNumber.Status.Status = "Failed"
		quantumRandomNumber.Status.Error = err.Error()
		_ = r.Status().Update(ctx, quantumRandomNumber)
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
func (r *QuantumRandomNumberReconciler) CreateOrUpdateSecret(quantumRandomNumber *qubeseciov1.QuantumRandomNumber, ctx context.Context) error {
	// Setup logger
	log := log.FromContext(ctx)

	secretName := quantumRandomNumber.Spec.SecretName
	if secretName == "" {
		secretName = quantumRandomNumber.Name
	}

	// Create Secret object
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: quantumRandomNumber.Namespace,
		},
	}

	// Get Secret object
	err := r.Get(ctx, client.ObjectKey{Namespace: secret.Namespace, Name: secret.Name}, secret)
	if err != nil && client.IgnoreNotFound(err) != nil {
		return err
	}

	// If Secret already exists, update status to Success
	if err == nil {
		if quantumRandomNumber.Status.Status != "Success" {
			now := metav1.Now()
			quantumRandomNumber.Status.Status = "Success"
			quantumRandomNumber.Status.RandomNumberReference = &qubeseciov1.ObjectReference{
				Name:      secretName,
				Namespace: quantumRandomNumber.Namespace,
			}
			quantumRandomNumber.Status.LastUpdateTime = &now
			quantumRandomNumber.Status.Error = ""
			_ = r.Status().Update(ctx, quantumRandomNumber)
		}
		return nil
	}

	// If Secret doesn't exist, create it
	secret, shannonEntropy := r.GenerateRandomNumberSecret(quantumRandomNumber, secretName, ctx)

	// Create Secret
	err = r.Create(ctx, secret)
	if err != nil {
		return err
	}
	log.Info("Created Secret")

	err = r.UpdateStatus(quantumRandomNumber, ctx, shannonEntropy)
	if err != nil {
		log.Error(err, "Create: Failed to Update Status")
		return err
	}

	// Update status to Success
	now := metav1.Now()
	quantumRandomNumber.Status.Status = "Success"
	quantumRandomNumber.Status.RandomNumberReference = &qubeseciov1.ObjectReference{
		Name:      secretName,
		Namespace: quantumRandomNumber.Namespace,
	}
	quantumRandomNumber.Status.LastUpdateTime = &now
	quantumRandomNumber.Status.Error = ""
	_ = r.Status().Update(ctx, quantumRandomNumber)

	return nil
}

// generate random number secret
func (r *QuantumRandomNumberReconciler) GenerateRandomNumberSecret(quantumRandomNumber *qubeseciov1.QuantumRandomNumber, secretName string, ctx context.Context) (*corev1.Secret, float64) {
	// Setup logger
	log := log.FromContext(ctx)

	// Set algorithm for quantum random number
	oqs.RandomBytesSwitchAlgorithm(quantumRandomNumber.Spec.Algorithm)

	// Generate quantum random number
	randomNumber := oqs.RandomBytes(quantumRandomNumber.Spec.Bytes)

	// Calculate Shannon Entropy
	shannonEntropy := shannonentropy.ShannonEntropy(randomNumber)

	// Create Secret object with quantum random number
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: quantumRandomNumber.Namespace,
		},
		Data: map[string][]byte{
			"quantumrandomnumber": randomNumber,
		},
	}

	// Set owner reference to QuantumRandomNumber for Secret
	err := ctrl.SetControllerReference(quantumRandomNumber, secret, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to Set Controller Reference")
	}

	return secret, shannonEntropy
}

// Update Status of QuantumRandomNumber
func (r *QuantumRandomNumberReconciler) UpdateStatus(quantumrandomnumber *qubeseciov1.QuantumRandomNumber, ctx context.Context, shannonEntropy float64) error {
	// Setup logger
	log := log.FromContext(ctx)

	// Update status of quantumrandomnumber to reflect the number of bytes of key material generated
	now := metav1.Now()
	quantumrandomnumber.Status.Status = "Success"
	quantumrandomnumber.Status.Bytes = quantumrandomnumber.Spec.Bytes
	quantumrandomnumber.Status.Algorithm = quantumrandomnumber.Spec.Algorithm
	quantumrandomnumber.Status.Entropy = fmt.Sprintf("%.12f", shannonEntropy)
	quantumrandomnumber.Status.LastUpdateTime = &now
	quantumrandomnumber.Status.Error = ""
	err := r.Status().Update(ctx, quantumrandomnumber)
	if err != nil {
		return err
	}
	log.Info("Updated Quantum Random Number Status")

	return nil
}

// applyDefaults sets default values for QuantumRandomNumber
func (r *QuantumRandomNumberReconciler) applyDefaults(qrng *qubeseciov1.QuantumRandomNumber) {
	if qrng.Spec.Bytes == 0 {
		qrng.Spec.Bytes = 32
	}
	if qrng.Spec.Algorithm == "" {
		qrng.Spec.Algorithm = "system"
	}
}

// validateQuantumRandomNumber validates the QuantumRandomNumber resource
func (r *QuantumRandomNumberReconciler) validateQuantumRandomNumber(qrng *qubeseciov1.QuantumRandomNumber) error {
	// Validate name length (max 52 characters for DNS compatibility)
	if len(qrng.ObjectMeta.Name) > 52 {
		return fmt.Errorf("name must be no more than 52 characters, got %d", len(qrng.ObjectMeta.Name))
	}

	// Validate seed
	if err := r.validateSeed(qrng); err != nil {
		return err
	}

	return nil
}

// validateSeed validates the seed specification
func (r *QuantumRandomNumberReconciler) validateSeed(qrng *qubeseciov1.QuantumRandomNumber) error {
	// If seed is not set but seedURI is provided, fetch it
	if qrng.Spec.Seed == "" && qrng.Spec.SeedURI != "" {
		seed, err := r.getSeedFromURI(qrng.Spec.SeedURI)
		if err != nil {
			return fmt.Errorf("failed to get seed from URI: %w", err)
		}
		qrng.Spec.Seed = seed
	}

	// Validate seed bytes length (if provided)
	if qrng.Spec.Seed != "" {
		seedBytes := []byte(qrng.Spec.Seed)
		if len(seedBytes) < 48 {
			return fmt.Errorf("seed length is %d bytes, must be at least 48 bytes", len(seedBytes))
		}
	}

	return nil
}

// getSeedFromURI fetches seed from a URI and converts to base64
func (r *QuantumRandomNumberReconciler) getSeedFromURI(seedURI string) (string, error) {
	resp, err := http.Get(seedURI)
	if err != nil {
		return "", fmt.Errorf("failed to get seed from %s: %w", seedURI, err)
	}
	defer resp.Body.Close()

	hexSeed, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read seed from %s: %w", seedURI, err)
	}

	seedInBytes, err := hex.DecodeString(string(hexSeed))
	if err != nil {
		return "", fmt.Errorf("failed to decode seed from %s: %w", seedURI, err)
	}

	// Convert hex seed to base64
	base64Seed := base64.StdEncoding.EncodeToString(seedInBytes)
	return base64Seed, nil
}


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

package v1

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	validationutils "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var quantumrandomnumberlog = logf.Log.WithName("quantumrandomnumber-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *QuantumRandomNumber) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-qubesec-io-v1-quantumrandomnumber,mutating=true,failurePolicy=fail,sideEffects=None,groups=qubesec.io,resources=quantumrandomnumbers,verbs=create;update,versions=v1,name=mquantumrandomnumber.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &QuantumRandomNumber{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *QuantumRandomNumber) Default() {
	quantumrandomnumberlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	fmt.Println("Default")
	// if Bytes is not set, set it to 32
	if r.Spec.Bytes == 0 {
		r.Spec.Bytes = 32
	}

	// if Algorithm is not set, set it to NIST-KAT
	if r.Spec.Algorithm == "" {
		r.Spec.Algorithm = "NIST-KAT"
	}

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-qubesec-io-v1-quantumrandomnumber,mutating=false,failurePolicy=fail,sideEffects=None,groups=qubesec.io,resources=quantumrandomnumbers,verbs=create;update,versions=v1,name=vquantumrandomnumber.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &QuantumRandomNumber{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *QuantumRandomNumber) ValidateCreate() (admission.Warnings, error) {
	quantumrandomnumberlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	fmt.Println("ValidateCreate")
	return nil, r.validateQuantumRandomNumber()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *QuantumRandomNumber) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	quantumrandomnumberlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, r.validateQuantumRandomNumber()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *QuantumRandomNumber) ValidateDelete() (admission.Warnings, error) {
	quantumrandomnumberlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *QuantumRandomNumber) validateQuantumRandomNumber() error {
	var allErrs field.ErrorList
	if err := r.validateQuantumRandomNumberName(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateQuantumRandomNumberSpec(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "qubesec.io", Kind: "QuantumRandomNumber"},
		r.Name, allErrs)
}

func (r *QuantumRandomNumber) validateQuantumRandomNumberSpec() *field.Error {
	// The field helpers from the kubernetes API machinery help us return nicely
	// structured validation errors.
	return validateSeedBytes(
		r.Spec.Seed,
		field.NewPath("spec").Child("Seed"))
}

func validateSeedBytes(seed string, fldPath *field.Path) *field.Error {
	// convert string to bytes array
	seedBytes := []byte(seed)
	// it should be more than 48 bytes
	if len(seedBytes) < 48 {
		return field.Invalid(fldPath, seed, "must be more than 48 bytes")
	}

	return nil
}

func (r *QuantumRandomNumber) validateQuantumRandomNumberName() *field.Error {
	if len(r.ObjectMeta.Name) > validationutils.DNS1035LabelMaxLength-11 {
		// The job name length is 63 character like all Kubernetes objects
		// (which must fit in a DNS subdomain). The QuantumRandomNumber controller appends
		// a 11-character suffix to the QuantumRandomNumber (`-$TIMESTAMP`) when creating
		// a job. The job name length limit is 63 characters. Therefore QuantumRandomNumber
		// names must have length <= 63-11=52. If we don't validate this here,
		// then job creation will fail later.
		return field.Invalid(field.NewPath("metadata").Child("name"), r.Name, "must be no more than 52 characters")
	}
	return nil
}

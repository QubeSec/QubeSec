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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// QuantumSharedSecretSpec defines the desired state of QuantumSharedSecret
type QuantumSharedSecretSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// PublicKeyRef is a reference to a QuantumKEMKeyPair that contains the public key
	// +kubebuilder:validation:Required
	PublicKeyRef ObjectReference `json:"publicKeyRef"`

	// Algorithm is the KEM algorithm to use (e.g., Kyber1024, Kyber768)
	// +kubebuilder:validation:Required
	Algorithm string `json:"algorithm"`

	// SecretName is the name of the secret to store the shared secret in
	// +kubebuilder:validation:Optional
	SecretName string `json:"secretName,omitempty"`
}

// ObjectReference contains enough information to let you inspect or modify the referred object
type ObjectReference struct {
	// Name of the referent
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace of the referent; empty defaults to current namespace
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
}

// QuantumSharedSecretStatus defines the observed state of QuantumSharedSecret.
type QuantumSharedSecretStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Status of the shared secret derivation
	// +kubebuilder:validation:Enum=Pending;Success;Failed
	Status string `json:"status,omitempty"`

	// Ciphertext is the encapsulated ciphertext (hex-encoded)
	Ciphertext string `json:"ciphertext,omitempty"`

	// SharedSecretReference points to where the shared secret is stored
	SharedSecretReference *ObjectReference `json:"sharedSecretReference,omitempty"`

	// LastUpdateTime is when the shared secret was last derived
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`

	// Error message if derivation failed
	Error string `json:"error,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=qss
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Algorithm",type=string,JSONPath=`.spec.algorithm`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// QuantumSharedSecret is the Schema for deriving shared secrets from KEM public keys
type QuantumSharedSecret struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of QuantumSharedSecret
	// +required
	Spec QuantumSharedSecretSpec `json:"spec"`

	// status defines the observed state of QuantumSharedSecret
	// +optional
	Status QuantumSharedSecretStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// QuantumSharedSecretList contains a list of QuantumSharedSecret
type QuantumSharedSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []QuantumSharedSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QuantumSharedSecret{}, &QuantumSharedSecretList{})
}

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

// QuantumDecapsulateSecretSpec defines the desired state of QuantumDecapsulateSecret
type QuantumDecapsulateSecretSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// PrivateKeyRef is a reference to a QuantumKEMKeyPair that contains the private key
	// +kubebuilder:validation:Required
	PrivateKeyRef ObjectReference `json:"privateKeyRef"`

	// Ciphertext is the encapsulated ciphertext (hex-encoded) from the encapsulation process
	// This is typically obtained from a QuantumEncapsulateSecret's status
	// +kubebuilder:validation:Optional
	Ciphertext string `json:"ciphertext,omitempty"`

	// CiphertextRef optionally points to a QuantumEncapsulateSecret to read ciphertext from status
	// If provided and ciphertext is empty, the controller will fetch the ciphertext from that resource
	// +kubebuilder:validation:Optional
	CiphertextRef *ObjectReference `json:"ciphertextRef,omitempty"`

	// Algorithm is the KEM algorithm to use (e.g., Kyber1024, Kyber768)
	// Must match the algorithm used during encapsulation
	// +kubebuilder:validation:Required
	Algorithm string `json:"algorithm"`

	// SecretName is the name of the secret to store the decapsulated shared secret in
	// +kubebuilder:validation:Optional
	SecretName string `json:"secretName,omitempty"`
}

// QuantumDecapsulateSecretStatus defines the observed state of QuantumDecapsulateSecret.
type QuantumDecapsulateSecretStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Status of the shared secret decapsulation
	// +kubebuilder:validation:Enum=Pending;Success;Failed
	Status string `json:"status,omitempty"`

	// SharedSecretReference points to where the shared secret is stored
	SharedSecretReference *ObjectReference `json:"sharedSecretReference,omitempty"`

	// Fingerprint is the SHA256 hash of the shared secret (first 10 characters)
	Fingerprint string `json:"fingerprint,omitempty"`

	// LastUpdateTime is when the shared secret was last decapsulated
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`

	// Error message if decapsulation failed
	Error string `json:"error,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=qds
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Algorithm",type=string,JSONPath=`.spec.algorithm`
// +kubebuilder:printcolumn:name="Fingerprint",type=string,JSONPath=`.status.fingerprint`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// QuantumDecapsulateSecret is the Schema for decapsulating shared secrets from KEM private keys and ciphertext
type QuantumDecapsulateSecret struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of QuantumDecapsulateSecret
	// +required
	Spec QuantumDecapsulateSecretSpec `json:"spec"`

	// status defines the observed state of QuantumDecapsulateSecret
	// +optional
	Status QuantumDecapsulateSecretStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// QuantumDecapsulateSecretList contains a list of QuantumDecapsulateSecret
type QuantumDecapsulateSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuantumDecapsulateSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QuantumDecapsulateSecret{}, &QuantumDecapsulateSecretList{})
}

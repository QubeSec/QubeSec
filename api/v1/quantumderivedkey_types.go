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

// QuantumDerivedKeySpec defines the desired state of QuantumDerivedKey
type QuantumDerivedKeySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// SharedSecretRef is a reference to a QuantumSharedSecret that contains the shared secret
	// +kubebuilder:validation:Required
	SharedSecretRef ObjectReference `json:"sharedSecretRef"`

	// KeyType specifies the type of key to derive (e.g., AES-256, ChaCha20)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=AES-256;ChaCha20;HMAC-SHA256
	KeyType string `json:"keyType"`

	// Salt is optional salt for the HKDF derivation (hex-encoded)
	// +kubebuilder:validation:Optional
	Salt string `json:"salt,omitempty"`

	// Info is optional info string for the HKDF derivation (hex-encoded)
	// +kubebuilder:validation:Optional
	Info string `json:"info,omitempty"`

	// SecretName is the name of the secret to store the derived key in
	// +kubebuilder:validation:Optional
	SecretName string `json:"secretName,omitempty"`
}

// QuantumDerivedKeyStatus defines the observed state of QuantumDerivedKey
type QuantumDerivedKeyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Status of the key derivation
	// +kubebuilder:validation:Enum=Pending;Success;Failed
	Status string `json:"status,omitempty"`

	// DerivedKeyReference points to where the derived key is stored
	DerivedKeyReference *ObjectReference `json:"derivedKeyReference,omitempty"`

	// LastUpdateTime is when the key was last derived
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`

	// KeyFingerprint is a hash of the derived key for verification (hex-encoded)
	KeyFingerprint string `json:"keyFingerprint,omitempty"`

	// FingerprintHash is the first 8 characters of the key fingerprint for quick verification
	FingerprintHash string `json:"fingerprintHash,omitempty"`

	// UsedSalt is the salt that was used in the derivation (hex-encoded or empty if not used)
	UsedSalt string `json:"usedSalt,omitempty"`

	// UsedInfo is the info that was used in the derivation (hex-encoded or empty if not used)
	UsedInfo string `json:"usedInfo,omitempty"`

	// Error message if derivation failed
	Error string `json:"error,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=qdk
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="KeyType",type=string,JSONPath=`.spec.keyType`
// +kubebuilder:printcolumn:name="FingerprintHash",type=string,JSONPath=`.status.fingerprintHash`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// QuantumDerivedKey is the Schema for deriving cryptographic keys from shared secrets using HKDF
type QuantumDerivedKey struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of QuantumDerivedKey
	// +required
	Spec QuantumDerivedKeySpec `json:"spec"`

	// status defines the observed state of QuantumDerivedKey
	// +optional
	Status QuantumDerivedKeyStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// QuantumDerivedKeyList contains a list of QuantumDerivedKey
type QuantumDerivedKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []QuantumDerivedKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QuantumDerivedKey{}, &QuantumDerivedKeyList{})
}

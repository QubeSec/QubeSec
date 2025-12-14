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

// QuantumSignatureKeyPairSpec defines the desired state of QuantumSignatureKeyPair
type QuantumSignatureKeyPairSpec struct {
	// Algorithm selects the signature scheme to use.
	// Supports liboqs names (Dilithium2/3/5, Falcon512/1024, SPHINCS+) and NIST names (ML-DSA-44/65/87, SLH-DSA-SHA2-128f)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Dilithium2;Dilithium3;Dilithium5;Falcon512;Falcon1024;SPHINCS+-SHA2-128f-simple;ML-DSA-44;ML-DSA-65;ML-DSA-87;SLH-DSA-SHA2-128f;SLH-DSA-SHA2-256f;CRYSTALS-Dilithium2;CRYSTALS-Dilithium3;CRYSTALS-Dilithium5
	// +kubebuilder:default=Dilithium2
	Algorithm string `json:"algorithm"`

	// Optional name of the Secret to store public/private keys. Defaults to resource name.
	// +kubebuilder:validation:Optional
	SecretName string `json:"secretName,omitempty"`
}

// QuantumSignatureKeyPairStatus defines the observed state of QuantumSignatureKeyPair
type QuantumSignatureKeyPairStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Status of key generation
	// +kubebuilder:validation:Enum=Pending;Success;Failed
	Status string `json:"status,omitempty"`

	// KeyPairReference points to where the keys are stored
	KeyPairReference *ObjectReference `json:"keyPairReference,omitempty"`

	// LastUpdateTime is when the key pair was last generated
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`

	// PublicKeyFingerprint is a hash of the public key (hex-encoded)
	PublicKeyFingerprint string `json:"publicKeyFingerprint,omitempty"`

	// Error message if generation failed
	Error string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=qskp
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
//+kubebuilder:printcolumn:name="Algorithm",type=string,JSONPath=`.spec.algorithm`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// QuantumSignatureKeyPair is the Schema for the quantumsignaturekeypairs API
type QuantumSignatureKeyPair struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuantumSignatureKeyPairSpec   `json:"spec,omitempty"`
	Status QuantumSignatureKeyPairStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// QuantumSignatureKeyPairList contains a list of QuantumSignatureKeyPair
type QuantumSignatureKeyPairList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuantumSignatureKeyPair `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QuantumSignatureKeyPair{}, &QuantumSignatureKeyPairList{})
}

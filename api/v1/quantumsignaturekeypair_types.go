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
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of QuantumSignatureKeyPair. Edit quantumsignaturekeypair_types.go to remove/update
	Algorithm string `json:"algorithm,omitempty"`
	// Optional name of the Secret to store public/private keys. Defaults to resource name.
	SecretName string `json:"secretName,omitempty"`
}

// QuantumSignatureKeyPairStatus defines the observed state of QuantumSignatureKeyPair
type QuantumSignatureKeyPairStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=qskp
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Algorithm",type=string,JSONPath=`.spec.algorithm`

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

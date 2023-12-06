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

// QuantumKeyPairSpec defines the desired state of QuantumKeyPair
type QuantumKeyPairSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of QuantumKeyPair. Edit quantumkeypair_types.go to remove/update
	Algorithm string `json:"algorithm,omitempty"`
}

// QuantumKeyPairStatus defines the observed state of QuantumKeyPair
type QuantumKeyPairStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Algorithm",type=string,JSONPath=`.spec.algorithm`

// QuantumKeyPair is the Schema for the quantumkeypairs API
type QuantumKeyPair struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuantumKeyPairSpec   `json:"spec,omitempty"`
	Status QuantumKeyPairStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// QuantumKeyPairList contains a list of QuantumKeyPair
type QuantumKeyPairList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuantumKeyPair `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QuantumKeyPair{}, &QuantumKeyPairList{})
}

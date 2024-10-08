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

// QuantumCertificateSpec defines the desired state of QuantumCertificate
type QuantumCertificateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of QuantumCertificate. Edit quantumcertificate_types.go to remove/update
	Algorithm string `json:"algorithm,omitempty"`
	Domain    string `json:"domain,omitempty"`
	Days      int    `json:"days,omitempty"`
}

// QuantumCertificateStatus defines the observed state of QuantumCertificate
type QuantumCertificateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=qc
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Algorithm",type=string,JSONPath=`.spec.algorithm`
//+kubebuilder:printcolumn:name="Domain",type=string,JSONPath=`.spec.domain`
//+kubebuilder:printcolumn:name="Days",type=number,JSONPath=`.spec.days`

// QuantumCertificate is the Schema for the quantumcertificates API
type QuantumCertificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuantumCertificateSpec   `json:"spec,omitempty"`
	Status QuantumCertificateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// QuantumCertificateList contains a list of QuantumCertificate
type QuantumCertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuantumCertificate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QuantumCertificate{}, &QuantumCertificateList{})
}

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

// QuantumRandomNumberSpec defines the desired state of QuantumRandomNumber
type QuantumRandomNumberSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of QuantumRandomNumber. Edit quantumrandomnumber_types.go to remove/update
	Bytes     int    `json:"bytes,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`
	Seed      string `json:"seed,omitempty"`
	SeedURI   string `json:"seedURI,omitempty"`
	// Optional name of the Secret to store the random number. Defaults to resource name.
	SecretName string `json:"secretName,omitempty"`
}

// QuantumRandomNumberStatus defines the observed state of QuantumRandomNumber
type QuantumRandomNumberStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Status of random number generation
	// +kubebuilder:validation:Enum=Pending;Success;Failed
	Status string `json:"status,omitempty"`

	// RandomNumberReference points to where the random data is stored
	RandomNumberReference *ObjectReference `json:"randomNumberReference,omitempty"`

	// LastUpdateTime is when the random number was last generated
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`

	Bytes     int    `json:"bytes,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`
	Entropy   string `json:"entropy,omitempty"`

	// Error message if generation failed
	Error string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=qrn;qrng
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
//+kubebuilder:printcolumn:name="Bytes",type=integer,JSONPath=`.status.bytes`
//+kubebuilder:printcolumn:name="Algorithm",type=string,JSONPath=`.status.algorithm`
//+kubebuilder:printcolumn:name="Entropy",type=string,JSONPath=`.status.entropy`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// QuantumRandomNumber is the Schema for the quantumrandomnumbers API
type QuantumRandomNumber struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuantumRandomNumberSpec   `json:"spec,omitempty"`
	Status QuantumRandomNumberStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// QuantumRandomNumberList contains a list of QuantumRandomNumber
type QuantumRandomNumberList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuantumRandomNumber `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QuantumRandomNumber{}, &QuantumRandomNumberList{})
}

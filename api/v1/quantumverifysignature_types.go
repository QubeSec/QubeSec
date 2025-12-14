package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// QuantumVerifySignatureSpec defines the desired state of QuantumVerifySignature.
type QuantumVerifySignatureSpec struct {
	// PublicKeyRef points to a QuantumSignatureKeyPair secret containing the public key.
	// +kubebuilder:validation:Required
	PublicKeyRef ObjectReference `json:"publicKeyRef"`

	// MessageRef points to a Secret that holds the message bytes under messageKey (default: "message").
	// +kubebuilder:validation:Required
	MessageRef ObjectReference `json:"messageRef"`

	// SignatureRef points to a Secret that holds the signature bytes under signatureKey (default: "signature").
	// +kubebuilder:validation:Required
	SignatureRef ObjectReference `json:"signatureRef"`

	// Algorithm selects the signature scheme to use.
	// Supports liboqs names (Dilithium2/3/5, Falcon512/1024, SPHINCS+) and NIST names (ML-DSA-44/65/87, SLH-DSA-SHA2-128f)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Dilithium2;Dilithium3;Dilithium5;Falcon512;Falcon1024;SPHINCS+-SHA2-128f-simple;ML-DSA-44;ML-DSA-65;ML-DSA-87;SLH-DSA-SHA2-128f;SLH-DSA-SHA2-256f;CRYSTALS-Dilithium2;CRYSTALS-Dilithium3;CRYSTALS-Dilithium5
	// +kubebuilder:default=Dilithium2
	Algorithm string `json:"algorithm"`

	// MessageKey selects the key in MessageRef data that contains the message bytes (default: "message").
	// +kubebuilder:validation:Optional
	MessageKey string `json:"messageKey,omitempty"`

	// SignatureKey selects the key that contains the signature bytes (default: "signature").
	// +kubebuilder:validation:Optional
	SignatureKey string `json:"signatureKey,omitempty"`
}

// QuantumVerifySignatureStatus defines the observed state of QuantumVerifySignature.
type QuantumVerifySignatureStatus struct {
	// +kubebuilder:validation:Enum=Pending;Valid;Invalid;Failed
	Status string `json:"status,omitempty"`

	// Verified is true when the signature is valid for the provided message.
	Verified bool `json:"verified,omitempty"`

	// MessageFingerprint is the SHA256 fingerprint of the verified message (first 10 hex chars).
	MessageFingerprint string `json:"messageFingerprint,omitempty"`

	// LastCheckedTime captures when the last verification attempt completed.
	LastCheckedTime *metav1.Time `json:"lastCheckedTime,omitempty"`

	// Error captures the failure reason.
	Error string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=qvs
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
//+kubebuilder:printcolumn:name="Algorithm",type=string,JSONPath=`.spec.algorithm`
//+kubebuilder:printcolumn:name="Verified",type=boolean,JSONPath=`.status.verified`
//+kubebuilder:printcolumn:name="Fingerprint",type=string,JSONPath=`.status.messageFingerprint`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// QuantumVerifySignature verifies a signature for message bytes using a referenced quantum-safe public key.
type QuantumVerifySignature struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuantumVerifySignatureSpec   `json:"spec"`
	Status QuantumVerifySignatureStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// QuantumVerifySignatureList contains a list of QuantumVerifySignature.
type QuantumVerifySignatureList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuantumVerifySignature `json:"items"`
}

func (in *QuantumVerifySignature) DeepCopyInto(out *QuantumVerifySignature) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status.Status = in.Status.Status
	out.Status.Verified = in.Status.Verified
	out.Status.MessageFingerprint = in.Status.MessageFingerprint
	out.Status.Error = in.Status.Error
	if in.Status.LastCheckedTime != nil {
		out.Status.LastCheckedTime = in.Status.LastCheckedTime.DeepCopy()
	}
}

func (in *QuantumVerifySignature) DeepCopy() *QuantumVerifySignature {
	if in == nil {
		return nil
	}
	out := new(QuantumVerifySignature)
	in.DeepCopyInto(out)
	return out
}

func (in *QuantumVerifySignature) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *QuantumVerifySignatureList) DeepCopyInto(out *QuantumVerifySignatureList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]QuantumVerifySignature, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *QuantumVerifySignatureList) DeepCopy() *QuantumVerifySignatureList {
	if in == nil {
		return nil
	}
	out := new(QuantumVerifySignatureList)
	in.DeepCopyInto(out)
	return out
}

func (in *QuantumVerifySignatureList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&QuantumVerifySignature{}, &QuantumVerifySignatureList{})
}

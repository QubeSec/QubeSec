package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// QuantumSignMessageSpec defines the desired state of QuantumSignMessage.
type QuantumSignMessageSpec struct {
	// PrivateKeyRef points to a QuantumSignatureKeyPair secret containing the private key.
	// +kubebuilder:validation:Required
	PrivateKeyRef ObjectReference `json:"privateKeyRef"`

	// MessageRef points to a Secret that holds the message bytes under messageKey (default: "message").
	// +kubebuilder:validation:Required
	MessageRef ObjectReference `json:"messageRef"`

	// Algorithm selects the signature scheme to use.
	// Supports liboqs names (Dilithium2/3/5, Falcon512/1024, SPHINCS+) and NIST names (ML-DSA-44/65/87, SLH-DSA-SHA2-128f)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Dilithium2;Dilithium3;Dilithium5;Falcon512;Falcon1024;SPHINCS+-SHA2-128f-simple;ML-DSA-44;ML-DSA-65;ML-DSA-87;SLH-DSA-SHA2-128f;SLH-DSA-SHA2-256f;CRYSTALS-Dilithium2;CRYSTALS-Dilithium3;CRYSTALS-Dilithium5
	// +kubebuilder:default=Dilithium2
	Algorithm string `json:"algorithm"`

	// OutputSecretName optionally overrides the Secret name where the signature is stored.
	// Defaults to <resource-name>-signature when empty.
	// +kubebuilder:validation:Optional
	OutputSecretName string `json:"outputSecretName,omitempty"`

	// MessageKey selects the key in MessageRef data that contains the message bytes (default: "message").
	// +kubebuilder:validation:Optional
	MessageKey string `json:"messageKey,omitempty"`

	// SignatureKey selects the key used to write the signature into the output Secret (default: "signature").
	// +kubebuilder:validation:Optional
	SignatureKey string `json:"signatureKey,omitempty"`
}

// QuantumSignMessageStatus defines the observed state of QuantumSignMessage.
type QuantumSignMessageStatus struct {
	// +kubebuilder:validation:Enum=Pending;Success;Failed
	Status string `json:"status,omitempty"`

	// Signature contains the base64-encoded signature (also written to the output Secret).
	Signature string `json:"signature,omitempty"`

	// SignatureReference points to the Secret containing the signature output.
	SignatureReference *ObjectReference `json:"signatureReference,omitempty"`

	// MessageFingerprint is the SHA256 fingerprint of the signed message (first 10 hex chars).
	MessageFingerprint string `json:"messageFingerprint,omitempty"`

	// LastUpdateTime is when the signature was last produced.
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`

	// Error captures the failure reason.
	Error string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=qsm
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
//+kubebuilder:printcolumn:name="Algorithm",type=string,JSONPath=`.spec.algorithm`
//+kubebuilder:printcolumn:name="Fingerprint",type=string,JSONPath=`.status.messageFingerprint`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// QuantumSignMessage signs message bytes using a referenced quantum-safe key pair.
type QuantumSignMessage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuantumSignMessageSpec   `json:"spec"`
	Status QuantumSignMessageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// QuantumSignMessageList contains a list of QuantumSignMessage.
type QuantumSignMessageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuantumSignMessage `json:"items"`
}

func (in *QuantumSignMessage) DeepCopyInto(out *QuantumSignMessage) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	if in.Status.SignatureReference != nil {
		out.Status.SignatureReference = &ObjectReference{}
		*out.Status.SignatureReference = *in.Status.SignatureReference
	}
	if in.Status.LastUpdateTime != nil {
		out.Status.LastUpdateTime = in.Status.LastUpdateTime.DeepCopy()
	}
	out.Status.Status = in.Status.Status
	out.Status.Signature = in.Status.Signature
	out.Status.MessageFingerprint = in.Status.MessageFingerprint
	out.Status.Error = in.Status.Error
}

func (in *QuantumSignMessage) DeepCopy() *QuantumSignMessage {
	if in == nil {
		return nil
	}
	out := new(QuantumSignMessage)
	in.DeepCopyInto(out)
	return out
}

func (in *QuantumSignMessage) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *QuantumSignMessageList) DeepCopyInto(out *QuantumSignMessageList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]QuantumSignMessage, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *QuantumSignMessageList) DeepCopy() *QuantumSignMessageList {
	if in == nil {
		return nil
	}
	out := new(QuantumSignMessageList)
	in.DeepCopyInto(out)
	return out
}

func (in *QuantumSignMessageList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&QuantumSignMessage{}, &QuantumSignMessageList{})
}

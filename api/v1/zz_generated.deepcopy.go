//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QuantumKEMKeyPair) DeepCopyInto(out *QuantumKEMKeyPair) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QuantumKEMKeyPair.
func (in *QuantumKEMKeyPair) DeepCopy() *QuantumKEMKeyPair {
	if in == nil {
		return nil
	}
	out := new(QuantumKEMKeyPair)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *QuantumKEMKeyPair) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QuantumKEMKeyPairList) DeepCopyInto(out *QuantumKEMKeyPairList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]QuantumKEMKeyPair, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QuantumKEMKeyPairList.
func (in *QuantumKEMKeyPairList) DeepCopy() *QuantumKEMKeyPairList {
	if in == nil {
		return nil
	}
	out := new(QuantumKEMKeyPairList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *QuantumKEMKeyPairList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QuantumKEMKeyPairSpec) DeepCopyInto(out *QuantumKEMKeyPairSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QuantumKEMKeyPairSpec.
func (in *QuantumKEMKeyPairSpec) DeepCopy() *QuantumKEMKeyPairSpec {
	if in == nil {
		return nil
	}
	out := new(QuantumKEMKeyPairSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QuantumKEMKeyPairStatus) DeepCopyInto(out *QuantumKEMKeyPairStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QuantumKEMKeyPairStatus.
func (in *QuantumKEMKeyPairStatus) DeepCopy() *QuantumKEMKeyPairStatus {
	if in == nil {
		return nil
	}
	out := new(QuantumKEMKeyPairStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QuantumRandomNumber) DeepCopyInto(out *QuantumRandomNumber) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QuantumRandomNumber.
func (in *QuantumRandomNumber) DeepCopy() *QuantumRandomNumber {
	if in == nil {
		return nil
	}
	out := new(QuantumRandomNumber)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *QuantumRandomNumber) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QuantumRandomNumberList) DeepCopyInto(out *QuantumRandomNumberList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]QuantumRandomNumber, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QuantumRandomNumberList.
func (in *QuantumRandomNumberList) DeepCopy() *QuantumRandomNumberList {
	if in == nil {
		return nil
	}
	out := new(QuantumRandomNumberList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *QuantumRandomNumberList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QuantumRandomNumberSpec) DeepCopyInto(out *QuantumRandomNumberSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QuantumRandomNumberSpec.
func (in *QuantumRandomNumberSpec) DeepCopy() *QuantumRandomNumberSpec {
	if in == nil {
		return nil
	}
	out := new(QuantumRandomNumberSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QuantumRandomNumberStatus) DeepCopyInto(out *QuantumRandomNumberStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QuantumRandomNumberStatus.
func (in *QuantumRandomNumberStatus) DeepCopy() *QuantumRandomNumberStatus {
	if in == nil {
		return nil
	}
	out := new(QuantumRandomNumberStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QuantumSignatureKeyPair) DeepCopyInto(out *QuantumSignatureKeyPair) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QuantumSignatureKeyPair.
func (in *QuantumSignatureKeyPair) DeepCopy() *QuantumSignatureKeyPair {
	if in == nil {
		return nil
	}
	out := new(QuantumSignatureKeyPair)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *QuantumSignatureKeyPair) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QuantumSignatureKeyPairList) DeepCopyInto(out *QuantumSignatureKeyPairList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]QuantumSignatureKeyPair, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QuantumSignatureKeyPairList.
func (in *QuantumSignatureKeyPairList) DeepCopy() *QuantumSignatureKeyPairList {
	if in == nil {
		return nil
	}
	out := new(QuantumSignatureKeyPairList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *QuantumSignatureKeyPairList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QuantumSignatureKeyPairSpec) DeepCopyInto(out *QuantumSignatureKeyPairSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QuantumSignatureKeyPairSpec.
func (in *QuantumSignatureKeyPairSpec) DeepCopy() *QuantumSignatureKeyPairSpec {
	if in == nil {
		return nil
	}
	out := new(QuantumSignatureKeyPairSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QuantumSignatureKeyPairStatus) DeepCopyInto(out *QuantumSignatureKeyPairStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QuantumSignatureKeyPairStatus.
func (in *QuantumSignatureKeyPairStatus) DeepCopy() *QuantumSignatureKeyPairStatus {
	if in == nil {
		return nil
	}
	out := new(QuantumSignatureKeyPairStatus)
	in.DeepCopyInto(out)
	return out
}

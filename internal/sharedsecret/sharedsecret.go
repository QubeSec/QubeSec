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

package sharedsecret

import (
	"context"
	"encoding/pem"
	"fmt"

	"github.com/open-quantum-safe/liboqs-go/oqs"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// DeriveSharedSecret uses KEM to encapsulate and derive a shared secret
func DeriveSharedSecret(algorithm string, publicKeyPEM []byte, ctx context.Context) ([]byte, []byte, error) {
	log := log.FromContext(ctx)

	// Decode PEM block
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to decode PEM block")
	}

	publicKey := block.Bytes

	// Initialize KEM
	quantumKEM := oqs.KeyEncapsulation{}
	defer quantumKEM.Clean()

	err := quantumKEM.Init(algorithm, nil)
	if err != nil {
		log.Error(err, "Failed to initialize KEM")
		return nil, nil, err
	}

	// Encapsulate to derive shared secret
	ciphertext, sharedSecret, err := quantumKEM.EncapSecret(publicKey)
	if err != nil {
		log.Error(err, "Failed to encapsulate secret")
		return nil, nil, err
	}

	return ciphertext, sharedSecret, nil
}

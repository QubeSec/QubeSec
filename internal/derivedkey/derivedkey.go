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

package derivedkey

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"golang.org/x/crypto/hkdf"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// DeriveAES256Key derives an AES-256 key using HKDF from a shared secret
// If no salt is provided, an empty salt is used (deterministic behavior)
func DeriveAES256Key(sharedSecretHex string, salt []byte, info []byte, ctx context.Context) (string, error) {
	log := log.FromContext(ctx)

	// Decode shared secret from hex
	sharedSecret, err := hex.DecodeString(sharedSecretHex)
	if err != nil {
		log.Error(err, "Failed to decode shared secret")
		return "", err
	}

	// Use HKDF to derive AES-256 key (32 bytes)
	// Note: If salt is nil/empty, HKDF uses an empty salt for deterministic derivation
	hash := sha256.New
	hkdf := hkdf.New(hash, sharedSecret, salt, info)

	derivedKey := make([]byte, 32) // 32 bytes = 256 bits for AES-256
	if _, err := io.ReadFull(hkdf, derivedKey); err != nil {
		log.Error(err, "Failed to derive key using HKDF")
		return "", err
	}

	derivedKeyHex := hex.EncodeToString(derivedKey)
	return derivedKeyHex, nil
}

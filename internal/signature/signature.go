package signature

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"

	"github.com/open-quantum-safe/liboqs-go/oqs"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// SignMessage signs a message using the provided private key and algorithm.
func SignMessage(algorithm string, privateKeyPEM []byte, message []byte, ctx context.Context) ([]byte, error) {
	log := log.FromContext(ctx)

	// Decode PEM block
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey := block.Bytes

	// Initialize signature scheme
	signer := oqs.Signature{}
	defer signer.Clean()

	err := signer.Init(algorithm, privateKey)
	if err != nil {
		log.Error(err, "Failed to initialize signature scheme")
		return nil, err
	}

	// Sign the message
	signature, err := signer.Sign(message)
	if err != nil {
		log.Error(err, "Failed to sign message")
		return nil, err
	}

	return signature, nil
}

// VerifySignature verifies a message signature using the provided public key and algorithm.
func VerifySignature(algorithm string, publicKeyPEM []byte, message []byte, signature []byte, ctx context.Context) (bool, error) {
	log := log.FromContext(ctx)

	// Decode PEM block
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return false, fmt.Errorf("failed to decode PEM block")
	}

	publicKey := block.Bytes

	// Initialize signature scheme
	verifier := oqs.Signature{}
	defer verifier.Clean()

	err := verifier.Init(algorithm, nil)
	if err != nil {
		log.Error(err, "Failed to initialize signature scheme")
		return false, err
	}

	// Verify the signature
	valid, err := verifier.Verify(message, signature, publicKey)
	if err != nil {
		log.Error(err, "Failed to verify signature")
		return false, err
	}

	return valid, nil
}

// MessageFingerprint computes the SHA256 fingerprint of a message and returns the first 10 hex chars.
func MessageFingerprint(message []byte) string {
	hash := sha256.Sum256(message)
	return hex.EncodeToString(hash[:])[:10]
}

// EncodeSignatureBase64 encodes the signature to base64.
func EncodeSignatureBase64(signature []byte) string {
	return base64.StdEncoding.EncodeToString(signature)
}

// DecodeSignatureBase64 decodes a base64-encoded signature.
func DecodeSignatureBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

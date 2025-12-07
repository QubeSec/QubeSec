package keypair

import (
	"bytes"
	"context"
	"encoding/pem"

	"github.com/open-quantum-safe/liboqs-go/oqs"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func GenerateKEMKeyPair(algorithm string, ctx context.Context) (string, string) {
	log := log.FromContext(ctx)

	quantumKeys := oqs.KeyEncapsulation{}
	defer quantumKeys.Clean()

	// Initialize liboqs-go
	err := quantumKeys.Init(algorithm, nil)
	if err != nil {
		log.Error(err, "Failed to initialize liboqs-go")
	}

	// Generate key pair
	quantumPublicKey, err := quantumKeys.GenerateKeyPair()
	if err != nil {
		log.Error(err, "Failed to generate key pair")
	}

	// Export private key
	quantumPrivateKey := quantumKeys.ExportSecretKey()

	return generatePEMBlock(quantumPublicKey, quantumPrivateKey, algorithm, ctx)
}

func GenerateSIGKeyPair(algorithm string, ctx context.Context) (string, string) {
	log := log.FromContext(ctx)

	quantumKeys := oqs.Signature{}
	defer quantumKeys.Clean()

	// Initialize liboqs-go
	err := quantumKeys.Init(algorithm, nil)
	if err != nil {
		log.Error(err, "Failed to initialize liboqs-go")
	}

	// Generate key pair
	quantumPublicKey, err := quantumKeys.GenerateKeyPair()
	if err != nil {
		log.Error(err, "Failed to generate key pair")
	}

	// Export private key
	quantumPrivateKey := quantumKeys.ExportSecretKey()

	return generatePEMBlock(quantumPublicKey, quantumPrivateKey, algorithm, ctx)
}

func generatePEMBlock(publicKey []byte, privateKey []byte, algorithm string, ctx context.Context) (string, string) {
	log := log.FromContext(ctx)

	// Generate PEM block
	publicKeyBlock := &pem.Block{
		Type:  algorithm + " PUBLIC KEY",
		Bytes: publicKey,
	}

	// Encode public key
	var publicKeyRow bytes.Buffer
	err := pem.Encode(&publicKeyRow, publicKeyBlock)
	if err != nil {
		log.Error(err, "Failed to encode public key")
	}

	// Generate PEM block
	privateKeyBlock := &pem.Block{
		Type:  algorithm + " SECRET KEY",
		Bytes: privateKey,
	}

	// Encode private key
	var privateKeyRow bytes.Buffer
	err = pem.Encode(&privateKeyRow, privateKeyBlock)
	if err != nil {
		log.Error(err, "Failed to encode private key")
	}

	// Return PEM encoded keys as strings
	return publicKeyRow.String(), privateKeyRow.String()
}

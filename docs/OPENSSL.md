# OpenSSL Commands for Quantum-Safe Cryptography

This guide provides OpenSSL commands for working with quantum-safe algorithms using the OQS provider (liboqs integration).

## Prerequisites

- OpenSSL 3.x or later
- OQS Provider (`oqsprovider`) installed
- liboqs library compiled

## Listing Algorithms

### List All Algorithms

```bash
openssl list -all-algorithms -provider oqsprovider
```

Shows all algorithms available through the OQS provider including quantum-safe signature schemes.

### List Signature Algorithms Only

```bash
openssl list -signature-algorithms -provider oqsprovider
```

Lists all signature/signing algorithms supported by the OQS provider (Dilithium, Falcon, SPHINCS+, etc.).

### List Key Exchange Algorithms (KEMs)

```bash
openssl list -kem-algorithms -provider oqsprovider
```

Lists key encapsulation mechanism algorithms available for key agreement.

### List All Provider Information

```bash
openssl list -providers
```

Shows all loaded providers (including oqsprovider if installed).

### List Detailed Provider Info

```bash
openssl provider -provider oqsprovider
```

Displays capabilities and algorithms available in the oqsprovider.

## Signature Algorithms Available

### Dilithium Variants

```bash
# List Dilithium signature algorithms
openssl list -signature-algorithms -provider oqsprovider | grep -i dilithium
```

Available variants:
- `Dilithium2` / `ML-DSA-44` (AES variant)
- `Dilithium3` / `ML-DSA-65` (AES variant)
- `Dilithium5` / `ML-DSA-87` (AES variant)

### Falcon Variants

```bash
# List Falcon signature algorithms
openssl list -signature-algorithms -provider oqsprovider | grep -i falcon
```

Available variants:
- `Falcon512` (512-bit security)
- `Falcon1024` (1024-bit security)

### SPHINCS+ Variants

```bash
# List SPHINCS+ stateless signature algorithms
openssl list -signature-algorithms -provider oqsprovider | grep -i sphincs
```

Available variants:
- `SPHINCS+-SHA2-128f-simple`
- `SPHINCS+-SHA2-128s-simple`
- `SPHINCS+-SHA2-192f-simple`
- `SPHINCS+-SHA2-192s-simple`
- `SPHINCS+-SHA2-256f-simple`
- `SPHINCS+-SHA2-256s-simple`

## Key Generation

### Generate Dilithium2 Keypair

```bash
openssl genpkey -algorithm Dilithium2 -out dilithium2_key.pem
```

Generates a new Dilithium2 private key.

### Generate Dilithium3 Keypair

```bash
openssl genpkey -algorithm Dilithium3 -out dilithium3_key.pem
```

Generates a new Dilithium3 private key.

### Generate Dilithium5 Keypair

```bash
openssl genpkey -algorithm Dilithium5 -out dilithium5_key.pem
```

Generates a new Dilithium5 private key (highest security).

### Generate Falcon512 Keypair

```bash
openssl genpkey -algorithm Falcon512 -out falcon512_key.pem
```

Generates a new Falcon512 private key.

### Generate Falcon1024 Keypair

```bash
openssl genpkey -algorithm Falcon1024 -out falcon1024_key.pem
```

Generates a new Falcon1024 private key (highest security).

### Generate SPHINCS+ Key

```bash
openssl genpkey -algorithm SPHINCS+-SHA2-128f-simple -out sphincs_key.pem
```

Generates a new SPHINCS+ private key with SHA2 and 128-bit security level.

### Using Provider with Key Generation

```bash
openssl genpkey -provider oqsprovider -algorithm Dilithium3 -out key.pem
```

Explicitly specifies the oqsprovider when generating keys.

## Extracting Public Keys

### Extract Public Key from Private Key

```bash
openssl pkey -in dilithium3_key.pem -pubout -out dilithium3_pub.pem
```

Extracts the public key portion from a private key file.

### Extract with Provider

```bash
openssl pkey -provider oqsprovider -in dilithium3_key.pem -pubout -out dilithium3_pub.pem
```

Explicitly uses the oqsprovider for key extraction.

### Display Public Key Text

```bash
openssl pkey -in dilithium3_pub.pem -pubin -text -noout
```

Displays the public key in human-readable format.

## Signing Messages

### Sign a Message

```bash
# Create a message
echo "Hello Quantum World!" > message.txt

# Sign with Dilithium3
openssl dgst -sign dilithium3_key.pem -out signature.sig message.txt
```

Signs a message using the private key. Output is in binary format.

### Sign with Specific Algorithm

```bash
openssl dgst -sign dilithium3_key.pem -out signature.sig message.txt
```

The algorithm is automatically determined from the key type.

### Sign and Output Base64

```bash
openssl dgst -sign dilithium3_key.pem message.txt | base64 > signature.b64
```

Signs and encodes the signature in Base64 for easier transport/storage.

### Verify Signature File

```bash
openssl dgst -verify dilithium3_pub.pem -signature signature.sig message.txt
```

Verifies a message signature using the public key.

### Using Explicit Provider

```bash
openssl dgst -provider oqsprovider -sign dilithium3_key.pem -out signature.sig message.txt
```

Explicitly specifies the oqsprovider for signing operations.

## Signature Verification

### Verify Message Signature

```bash
openssl dgst -verify dilithium3_pub.pem -signature signature.sig message.txt
```

Returns "Verified OK" or "Verification Failure".

### Verify with Provider

```bash
openssl dgst -provider oqsprovider -verify dilithium3_pub.pem -signature signature.sig message.txt
```

Explicitly uses the oqsprovider for verification.

### Verify Multiple Algorithms

```bash
# Dilithium
openssl dgst -verify dilithium3_pub.pem -signature signature.sig message.txt

# Falcon
openssl dgst -verify falcon512_pub.pem -signature falcon_sig.sig message.txt

# SPHINCS+
openssl dgst -verify sphincs_pub.pem -signature sphincs_sig.sig message.txt
```

Each algorithm uses its own public key and signature file.

## Key Information Commands

### Display Key Information

```bash
openssl pkey -in dilithium3_key.pem -text -noout
```

Shows detailed information about the private key (use with caution - reveals sensitive data).

### Get Key Type

```bash
openssl pkey -in dilithium3_key.pem -noout -text | head -1
```

Displays the key type/algorithm.

### Check Key Format

```bash
openssl pkey -in dilithium3_key.pem -check
```

Validates the key file format and structure.

### Public Key Info

```bash
openssl pkey -in dilithium3_pub.pem -pubin -text -noout
```

Displays public key information without sensitive data.

## Working with Different Signature Schemes

### Complete Dilithium2 Workflow

```bash
# 1. Generate keypair
openssl genpkey -algorithm Dilithium2 -out dil2_key.pem

# 2. Extract public key
openssl pkey -in dil2_key.pem -pubout -out dil2_pub.pem

# 3. Create message
echo "Dilithium2 Test" > msg.txt

# 4. Sign
openssl dgst -sign dil2_key.pem -out dil2_sig.sig msg.txt

# 5. Verify
openssl dgst -verify dil2_pub.pem -signature dil2_sig.sig msg.txt
```

Complete workflow from key generation to signature verification.

### Complete Falcon Workflow

```bash
# 1. Generate keypair
openssl genpkey -algorithm Falcon512 -out falcon_key.pem

# 2. Extract public key
openssl pkey -in falcon_key.pem -pubout -out falcon_pub.pem

# 3. Create message
echo "Falcon Test" > msg.txt

# 4. Sign
openssl dgst -sign falcon_key.pem -out falcon_sig.sig msg.txt

# 5. Verify
openssl dgst -verify falcon_pub.pem -signature falcon_sig.sig msg.txt
```

Complete Falcon512 signing workflow.

### Complete SPHINCS+ Workflow

```bash
# 1. Generate keypair
openssl genpkey -algorithm SPHINCS+-SHA2-128f-simple -out sphincs_key.pem

# 2. Extract public key
openssl pkey -in sphincs_key.pem -pubout -out sphincs_pub.pem

# 3. Create message
echo "SPHINCS+ Test" > msg.txt

# 4. Sign
openssl dgst -sign sphincs_key.pem -out sphincs_sig.sig msg.txt

# 5. Verify
openssl dgst -verify sphincs_pub.pem -signature sphincs_sig.sig msg.txt
```

Complete SPHINCS+ stateless signature workflow.

## Algorithm Comparison Commands

### Compare Key Sizes

```bash
# Dilithium keys
openssl pkey -in dilithium3_key.pem -text -noout | grep -i "size\|bits"

# Falcon keys
openssl pkey -in falcon512_key.pem -text -noout | grep -i "size\|bits"

# SPHINCS+ keys
openssl pkey -in sphincs_key.pem -text -noout | grep -i "size\|bits"
```

Compare key sizes across different algorithms.

### Check Signature Sizes

```bash
# Get file sizes in bytes
ls -l dilithium3_sig.sig falcon512_sig.sig sphincs_sig.sig

# Alternative with wc
wc -c dilithium3_sig.sig falcon512_sig.sig sphincs_sig.sig
```

Compare signature size differences across algorithms.

## NIST Algorithm Name Mapping

OpenSSL may support both traditional names and NIST standard names:

```bash
# List algorithms that might support NIST naming
openssl list -signature-algorithms -provider oqsprovider

# These are equivalent:
openssl genpkey -algorithm Dilithium2
openssl genpkey -algorithm ML-DSA-44

openssl genpkey -algorithm Dilithium3
openssl genpkey -algorithm ML-DSA-65

openssl genpkey -algorithm Dilithium5
openssl genpkey -algorithm ML-DSA-87
```

Check your OpenSSL version for NIST name support.

## Batch Operations

### Generate Multiple Keypairs

```bash
#!/bin/bash
for algo in Dilithium2 Dilithium3 Dilithium5 Falcon512 Falcon1024; do
  openssl genpkey -algorithm $algo -out ${algo}_key.pem
  openssl pkey -in ${algo}_key.pem -pubout -out ${algo}_pub.pem
  echo "Generated $algo keypair"
done
```

Script to generate multiple algorithm keypairs.

### Test All Algorithms

```bash
#!/bin/bash
for algo in Dilithium2 Dilithium3 Dilithium5 Falcon512 Falcon1024; do
  echo "Testing $algo..."
  
  # Generate
  openssl genpkey -algorithm $algo -out key.pem
  openssl pkey -in key.pem -pubout -out pub.pem
  
  # Sign & Verify
  echo "Test Message" | openssl dgst -sign key.pem | openssl dgst -verify pub.pem
  
  rm key.pem pub.pem
done
```

Script to test all supported algorithms.

## Troubleshooting

### Provider Not Found Error

```bash
# Check if oqsprovider is installed
openssl list -providers

# Explicitly load provider
openssl list -all-algorithms -provider oqsprovider

# Check provider path
openssl list -providers -details
```

If oqsprovider is not listed, install liboqs-provider.

### Algorithm Not Supported

```bash
# Verify algorithm is available
openssl list -signature-algorithms -provider oqsprovider | grep -i Dilithium

# Check OpenSSL version
openssl version
```

Ensure OpenSSL 3.x and oqsprovider are properly configured.

### Key Format Issues

```bash
# Check key format
file dilithium3_key.pem

# Validate key structure
openssl pkey -in dilithium3_key.pem -check -text

# Try with explicit provider
openssl pkey -provider oqsprovider -in dilithium3_key.pem -check
```

Ensures keys are in PEM format and properly structured.

## Integration with QubeSec

These OpenSSL commands can be used to:

1. **Pre-generate keypairs** for QuantumSignatureKeyPair resources
2. **Manually test** signature operations before deploying to Kubernetes
3. **Verify** that signatures created by QubeSec controllers are valid
4. **Debug** algorithm-specific issues
5. **Generate** test data for CI/CD pipelines

### Example: Prepare Keys for QubeSec

```bash
# Generate key
openssl genpkey -algorithm Dilithium3 -out my-key.pem

# Extract public key
openssl pkey -in my-key.pem -pubout -out my-key-pub.pem

# Create Kubernetes secret
kubectl create secret generic my-quantum-keys \
  --from-file=private-key=my-key.pem \
  --from-file=public-key=my-key-pub.pem

# Reference in QuantumSignatureKeyPair CRD
kubectl apply -f - <<EOF
apiVersion: qubesec.io/v1
kind: QuantumSignatureKeyPair
metadata:
  name: my-keys
spec:
  algorithm: Dilithium3
  keySecret: my-quantum-keys
EOF
```

Prepare keys using OpenSSL for use with QubeSec.

## Additional Resources

- [liboqs Documentation](https://liboqs.org/)
- [OpenSSL 3.x Documentation](https://www.openssl.org/docs/man3.0/)
- [NIST Post-Quantum Cryptography](https://csrc.nist.gov/projects/post-quantum-cryptography/)
- [QubeSec Documentation](../README.md)

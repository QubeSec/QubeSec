# Quantum Signature Implementation Guide

## Overview

This document describes the quantum-safe digital signature framework added to QubeSec. It includes three main components:

1. **QuantumSignatureKeyPair** - Generates and stores quantum-safe keypairs
2. **QuantumSignMessage** - Signs messages using a private key
3. **QuantumVerifySignature** - Verifies signatures using a public key

---

## 1. QuantumSignatureKeyPair CRD (Enhanced)

### Location
- **Types**: [api/v1/quantumsignaturekeypair_types.go](api/v1/quantumsignaturekeypair_types.go)
- **Controller**: [internal/controller/quantumsignaturekeypair_controller.go](internal/controller/quantumsignaturekeypair_controller.go)

### Spec Fields
| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `algorithm` | ✅ Yes | Enum | Signature algorithm (Dilithium2, Dilithium3, Dilithium5, Falcon512, Falcon1024, SPHINCS+-SHA2-128f-simple) |
| `secretName` | ❌ No | String | Secret name for key storage (defaults to resource name) |

### Status Fields
| Field | Type | Description |
|-------|------|-------------|
| `status` | String (Pending/Success/Failed) | Keypair generation status |
| `keyPairReference` | ObjectReference | Reference to Secret containing keys |
| `publicKeyFingerprint` | String | SHA256 fingerprint of public key (first 10 hex chars) |
| `lastUpdateTime` | Time | Timestamp of last generation |
| `error` | String | Error message if generation failed |

### Example
```yaml
apiVersion: qubesec.io/v1
kind: QuantumSignatureKeyPair
metadata:
  name: example-sig-keypair
spec:
  algorithm: Dilithium2
  secretName: my-sig-keys
```

### Enhancements Made
- ✅ Algorithm is now **required** with **enum validation**
- ✅ **Error handling**: Generation errors are caught and reported in status
- ✅ **PublicKeyFingerprint**: Now computed and stored
- ✅ **Status updates**: Proper error handling for status updates
- ✅ **Secret validation**: Checks for required keys when secret already exists

---

## 2. QuantumSignMessage CRD (New)

### Location
- **Types**: [api/v1/quantumsignmessage_types.go](api/v1/quantumsignmessage_types.go)
- **Controller**: [internal/controller/quantumsignmessage_controller.go](internal/controller/quantumsignmessage_controller.go)
- **Sample**: [config/samples/_v1_quantumsignmessage.yaml](config/samples/_v1_quantumsignmessage.yaml)

### Spec Fields
| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `algorithm` | ✅ Yes | Enum | Signature scheme (same as keypair algorithms) |
| `privateKeyRef` | ✅ Yes | ObjectReference | Reference to QuantumSignatureKeyPair secret |
| `messageRef` | ✅ Yes | ObjectReference | Secret containing message bytes |
| `outputSecretName` | ❌ No | String | Output secret name (defaults to `<name>-signature`) |
| `messageKey` | ❌ No | String | Key in messageRef (default: "message") |
| `signatureKey` | ❌ No | String | Key for storing signature (default: "signature") |

### Status Fields
| Field | Type | Description |
|-------|------|-------------|
| `status` | String (Pending/Success/Failed) | Signing operation status |
| `signature` | String | Base64-encoded signature |
| `signatureReference` | ObjectReference | Reference to output Secret |
| `messageFingerprint` | String | SHA256 fingerprint of message (first 10 hex chars) |
| `lastUpdateTime` | Time | Timestamp of signature creation |
| `error` | String | Error message if signing failed |

### Example
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sample-message
data:
  message: "SGVsbG8sIFF1YW50dW0gV29ybGQh"  # base64: "Hello, Quantum World!"
---
apiVersion: qubesec.io/v1
kind: QuantumSignMessage
metadata:
  name: sign-example
spec:
  algorithm: Dilithium2
  privateKeyRef:
    name: example-sig-keypair
  messageRef:
    name: sample-message
  outputSecretName: signed-output
```

### Controller Behavior
1. Validates algorithm is specified
2. Fetches QuantumSignatureKeyPair and extracts private key from Secret
3. Fetches message from referenced Secret
4. Calls `signature.SignMessage()` to generate signature
5. Creates or updates output Secret with signature bytes
6. Updates status with base64-encoded signature and message fingerprint

---

## 3. QuantumVerifySignature CRD (New)

### Location
- **Types**: [api/v1/quantumverifysignature_types.go](api/v1/quantumverifysignature_types.go)
- **Controller**: [internal/controller/quantumverifysignature_controller.go](internal/controller/quantumverifysignature_controller.go)
- **Sample**: [config/samples/_v1_quantumverifysignature.yaml](config/samples/_v1_quantumverifysignature.yaml)

### Spec Fields
| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `algorithm` | ✅ Yes | Enum | Signature scheme (must match key generation) |
| `publicKeyRef` | ✅ Yes | ObjectReference | Reference to QuantumSignatureKeyPair secret |
| `messageRef` | ✅ Yes | ObjectReference | Secret containing original message bytes |
| `signatureRef` | ✅ Yes | ObjectReference | Secret containing signature to verify |
| `messageKey` | ❌ No | String | Key in messageRef (default: "message") |
| `signatureKey` | ❌ No | String | Key in signatureRef (default: "signature") |

### Status Fields
| Field | Type | Description |
|-------|------|-------------|
| `status` | String (Pending/Valid/Invalid/Failed) | Verification result |
| `verified` | Boolean | True if signature is valid |
| `messageFingerprint` | String | SHA256 fingerprint of verified message (first 10 hex chars) |
| `lastCheckedTime` | Time | Timestamp of verification |
| `error` | String | Error message if verification failed |

### Example
```yaml
apiVersion: qubesec.io/v1
kind: QuantumVerifySignature
metadata:
  name: verify-example
spec:
  algorithm: Dilithium2
  publicKeyRef:
    name: example-sig-keypair
  messageRef:
    name: sample-message
  signatureRef:
    name: signed-output
  messageKey: message
  signatureKey: signature
```

### Controller Behavior
1. Validates algorithm is specified
2. Fetches QuantumSignatureKeyPair and extracts public key from Secret
3. Fetches message and signature from referenced Secrets
4. Calls `signature.VerifySignature()` to verify
5. Sets status to "Valid" or "Invalid" based on result
6. Updates messageFingerprint for audit trail

---

## 4. Signature Utility Package

### Location
[internal/signature/signature.go](internal/signature/signature.go)

### Functions

#### `SignMessage(algorithm, privateKeyPEM, message, ctx) -> (signature, error)`
Signs a message using a private key with the specified algorithm.
- Decodes PEM-formatted private key
- Initializes oqs.Signature with algorithm
- Calls `Sign()` on the message
- Returns raw signature bytes

#### `VerifySignature(algorithm, publicKeyPEM, message, signature, ctx) -> (valid, error)`
Verifies a message signature using a public key.
- Decodes PEM-formatted public key
- Initializes oqs.Signature with algorithm
- Calls `Verify()` with message, signature, and public key
- Returns true/false for validity

#### `MessageFingerprint(message) -> string`
Computes SHA256 hash of message and returns first 10 hex characters.

#### `EncodeSignatureBase64(signature) -> string`
Encodes signature bytes to base64 for status/storage.

#### `DecodeSignatureBase64(encoded) -> (bytes, error)`
Decodes base64-encoded signature back to bytes.

---

## 5. Supporting Changes

### Enhanced Keypair Generation ([internal/keypair/keypair.go](internal/keypair/keypair.go))

Both `GenerateKEMKeyPair()` and `GenerateSIGKeyPair()` now:
- Return error as third return value: `(publicKey, privateKey, error)`
- Properly handle and propagate errors from liboqs-go
- Support PEM encoding with error handling

### Controller Registration ([cmd/main.go](cmd/main.go))

Added reconcilers:
- `QuantumSignMessageReconciler`
- `QuantumVerifySignatureReconciler`

---

## 6. Security Best Practices

### Key Storage
- **Private keys** stored in Kubernetes Secrets (encrypted at rest with Secrets encryption)
- **Public keys** accessible via ObjectReference for verification
- Secrets owned by CRs for automatic cleanup via garbage collection

### Message Handling
- **Messages** stored in referenced Secrets (not inline in CRD)
- Large messages supported via Secret reference pattern
- Message fingerprints stored for audit trail

### Error Handling
- **Validation errors** reported in status.error field
- **Signature verification failures** marked as "Invalid" status
- No silent failures - all errors surface in status

### Algorithm Validation
- Algorithm enum prevents typos and unsupported schemes
- Algorithm mismatch during verification caught and reported

---

## 7. Workflow Example

### Step 1: Generate Keypair
```bash
kubectl apply -f - <<EOF
apiVersion: qubesec.io/v1
kind: QuantumSignatureKeyPair
metadata:
  name: my-keys
spec:
  algorithm: Dilithium2
EOF
```

### Step 2: Create Message Secret
```bash
kubectl create secret generic my-message \
  --from-literal=message="Hello, Quantum World!"
```

### Step 3: Sign Message
```bash
kubectl apply -f - <<EOF
apiVersion: qubesec.io/v1
kind: QuantumSignMessage
metadata:
  name: my-signature
spec:
  algorithm: Dilithium2
  privateKeyRef:
    name: my-keys
  messageRef:
    name: my-message
EOF

# Check status
kubectl get quantumsignmessage my-signature -o jsonpath='{.status}'
```

### Step 4: Verify Signature
```bash
kubectl apply -f - <<EOF
apiVersion: qubesec.io/v1
kind: QuantumVerifySignature
metadata:
  name: verify-my-sig
spec:
  algorithm: Dilithium2
  publicKeyRef:
    name: my-keys
  messageRef:
    name: my-message
  signatureRef:
    name: my-signature-signature
EOF

# Check result
kubectl get quantumverifysignature verify-my-sig -o jsonpath='{.status.verified}'
```

---

## 8. Supported Algorithms

All algorithms from liboqs-go, including:
- **Dilithium** (recommended): Dilithium2, Dilithium3, Dilithium5
- **Falcon**: Falcon512, Falcon1024
- **SPHINCS+**: SPHINCS+-SHA2-128f-simple (and variants)

---

## 9. Future Enhancements

Potential additions:
- [ ] Context string support for signature schemes that support it (SigWithCtx)
- [ ] Batch signing/verification operations
- [ ] Signature expiration and timestamping
- [ ] Hardware security module (HSM) integration for key storage
- [ ] Webhook validation for algorithm enum enforcement
- [ ] Metrics and audit logging enhancements

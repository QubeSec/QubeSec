# Quantum Signature - Usage Examples

## Complete Workflow Example

### Prerequisites
- Kubernetes cluster with QubeSec operator running
- `kubectl` configured to access cluster

---

## Example 1: Basic Sign & Verify Workflow

### Step 1: Create a QuantumSignatureKeyPair

```bash
kubectl apply -f - <<'EOF'
apiVersion: qubesec.io/v1
kind: QuantumSignatureKeyPair
metadata:
  name: my-signing-keys
  namespace: default
spec:
  algorithm: Dilithium2
  secretName: my-sig-secret
EOF

# Verify it was created
kubectl get quantumsignaturekeypair my-signing-keys
kubectl describe quantumsignaturekeypair my-signing-keys
```

**Expected Output:**
```
NAME                 STATUS    ALGORITHM    AGE
my-signing-keys      Success   Dilithium2   10s
```

### Step 2: Create a Secret with Your Message

```bash
kubectl create secret generic my-message \
  --from-literal=message="Hello, Quantum-Safe World!"

# Verify the secret
kubectl get secret my-message
kubectl get secret my-message -o jsonpath='{.data.message}' | base64 -d
```

### Step 3: Sign the Message

```bash
kubectl apply -f - <<'EOF'
apiVersion: qubesec.io/v1
kind: QuantumSignMessage
metadata:
  name: sign-my-message
  namespace: default
spec:
  algorithm: Dilithium2
  privateKeyRef:
    name: my-signing-keys
    namespace: default
  messageRef:
    name: my-message
    namespace: default
  outputSecretName: my-signature-output
  messageKey: message
  signatureKey: signature
EOF

# Check signing status
kubectl get quantumsignmessage sign-my-message
kubectl describe quantumsignmessage sign-my-message
```

**Expected Output:**
```
NAME                 STATUS     ALGORITHM    SIGNATURE              AGE
sign-my-message      Success    Dilithium2   QlKK3jX1dxD4Z8...    5s
```

### Step 4: Verify the Signature

```bash
kubectl apply -f - <<'EOF'
apiVersion: qubesec.io/v1
kind: QuantumVerifySignature
metadata:
  name: verify-my-signature
  namespace: default
spec:
  algorithm: Dilithium2
  publicKeyRef:
    name: my-signing-keys
    namespace: default
  messageRef:
    name: my-message
    namespace: default
  signatureRef:
    name: my-signature-output
    namespace: default
  messageKey: message
  signatureKey: signature
EOF

# Check verification result
kubectl get quantumverifysignature verify-my-signature
kubectl describe quantumverifysignature verify-my-signature
```

**Expected Output:**
```
NAME                        STATUS    VERIFIED    AGE
verify-my-signature         Valid     true        3s
```

---

## Example 2: Different Algorithms

### Sign with Falcon512

```bash
kubectl apply -f - <<'EOF'
apiVersion: qubesec.io/v1
kind: QuantumSignatureKeyPair
metadata:
  name: falcon-keys
spec:
  algorithm: Falcon512
EOF

kubectl apply -f - <<'EOF'
apiVersion: qubesec.io/v1
kind: QuantumSignMessage
metadata:
  name: sign-with-falcon
spec:
  algorithm: Falcon512
  privateKeyRef:
    name: falcon-keys
  messageRef:
    name: my-message
EOF

kubectl get quantumsignmessage sign-with-falcon -o jsonpath='{.status}'
```

### Sign with Dilithium5

```bash
kubectl apply -f - <<'EOF'
apiVersion: qubesec.io/v1
kind: QuantumSignatureKeyPair
metadata:
  name: dilithium5-keys
spec:
  algorithm: Dilithium5
EOF

kubectl apply -f - <<'EOF'
apiVersion: qubesec.io/v1
kind: QuantumSignMessage
metadata:
  name: sign-with-dilithium5
spec:
  algorithm: Dilithium5
  privateKeyRef:
    name: dilithium5-keys
  messageRef:
    name: my-message
EOF
```

---

## Example 3: Multiple Messages with Same Key

```bash
# Create different messages
kubectl create secret generic message-1 --from-literal=message="Message 1"
kubectl create secret generic message-2 --from-literal=message="Message 2"
kubectl create secret generic message-3 --from-literal=message="Message 3"

# Sign all with same key
for i in 1 2 3; do
  kubectl apply -f - <<EOF
apiVersion: qubesec.io/v1
kind: QuantumSignMessage
metadata:
  name: sign-message-$i
spec:
  algorithm: Dilithium2
  privateKeyRef:
    name: my-signing-keys
  messageRef:
    name: message-$i
  outputSecretName: signature-$i
EOF
done

# Verify all
for i in 1 2 3; do
  kubectl apply -f - <<EOF
apiVersion: qubesec.io/v1
kind: QuantumVerifySignature
metadata:
  name: verify-message-$i
spec:
  algorithm: Dilithium2
  publicKeyRef:
    name: my-signing-keys
  messageRef:
    name: message-$i
  signatureRef:
    name: signature-$i
EOF
done

# Check all results
kubectl get quantumverifysignature
```

---

## Example 4: Working with Binary Data

### Sign Binary Data (e.g., Container Image Digest)

```bash
# Create secret with binary image digest
kubectl create secret generic image-digest \
  --from-literal=digest="sha256:1234567890abcdef..."

# Sign the digest
kubectl apply -f - <<'EOF'
apiVersion: qubesec.io/v1
kind: QuantumSignMessage
metadata:
  name: sign-image-digest
spec:
  algorithm: Dilithium2
  privateKeyRef:
    name: my-signing-keys
  messageRef:
    name: image-digest
  outputSecretName: image-signature
  messageKey: digest
  signatureKey: image-signature
EOF

# Verify the signature
kubectl apply -f - <<'EOF'
apiVersion: qubesec.io/v1
kind: QuantumVerifySignature
metadata:
  name: verify-image-digest
spec:
  algorithm: Dilithium2
  publicKeyRef:
    name: my-signing-keys
  messageRef:
    name: image-digest
  signatureRef:
    name: image-signature
  messageKey: digest
  signatureKey: image-signature
EOF
```

---

## Example 5: Checking Fingerprints for Audit

```bash
# Get message fingerprint
kubectl get quantumsignmessage sign-my-message \
  -o jsonpath='{.status.messageFingerprint}'

# Get public key fingerprint
kubectl get quantumsignaturekeypair my-signing-keys \
  -o jsonpath='{.status.publicKeyFingerprint}'

# Get verification timestamp
kubectl get quantumverifysignature verify-my-signature \
  -o jsonpath='{.status.lastCheckedTime}'
```

---

## Example 6: Cross-Namespace References

### Sign message in different namespace

```bash
# Create message in one namespace
kubectl create namespace signing-ns
kubectl create secret generic message \
  --from-literal=message="Cross-namespace message" \
  -n signing-ns

# Sign using keys from default namespace
kubectl apply -f - <<'EOF'
apiVersion: qubesec.io/v1
kind: QuantumSignMessage
metadata:
  name: cross-ns-sign
  namespace: signing-ns
spec:
  algorithm: Dilithium2
  privateKeyRef:
    name: my-signing-keys
    namespace: default
  messageRef:
    name: message
    namespace: signing-ns
EOF

# Check status
kubectl get quantumsignmessage cross-ns-sign -n signing-ns
```

---

## Example 7: Error Handling

### Missing Required Fields

```bash
# This will fail - no algorithm specified
kubectl apply -f - <<'EOF'
apiVersion: qubesec.io/v1
kind: QuantumSignMessage
metadata:
  name: missing-algo
spec:
  privateKeyRef:
    name: my-signing-keys
  messageRef:
    name: my-message
EOF

# Check the error
kubectl describe quantumsignmessage missing-algo
# Status: Failed, Error: "spec.algorithm is required"
```

### Invalid Algorithm

```bash
# This will fail - enum validation
kubectl apply -f - <<'EOF'
apiVersion: qubesec.io/v1
kind: QuantumSignatureKeyPair
metadata:
  name: invalid-algo
spec:
  algorithm: InvalidAlgo
EOF

# Error from API server: spec.algorithm must be one of...
```

### Missing Message Secret

```bash
# This will fail - referenced secret doesn't exist
kubectl apply -f - <<'EOF'
apiVersion: qubesec.io/v1
kind: QuantumSignMessage
metadata:
  name: missing-message
spec:
  algorithm: Dilithium2
  privateKeyRef:
    name: my-signing-keys
  messageRef:
    name: nonexistent-secret
EOF

# Check the error
kubectl describe quantumsignmessage missing-message
# Status: Failed, Error: "Failed to get message secret: secrets \"nonexistent-secret\" not found"
```

---

## Example 8: Cleanup & Garbage Collection

```bash
# Delete message signing
kubectl delete quantumsignmessage sign-my-message

# The output secret is automatically deleted (owner reference)
kubectl get secret my-signature-output

# Delete key pair
kubectl delete quantumsignaturekeypair my-signing-keys

# The key secret is automatically deleted
kubectl get secret my-sig-secret
```

---

## Example 9: Checking Statuses in Detail

```bash
# Get all fields from QuantumSignMessage
kubectl get quantumsignmessage sign-my-message \
  -o jsonpath='{
    .status.status,
    .status.verified,
    .status.messageFingerprint,
    .status.signature,
    .status.lastUpdateTime
  }'

# Get all fields from QuantumVerifySignature
kubectl get quantumverifysignature verify-my-signature \
  -o jsonpath='{
    .status.status,
    .status.verified,
    .status.messageFingerprint,
    .status.lastCheckedTime
  }'
```

---

## Example 10: Continuous Integration Example

```bash
#!/bin/bash

# Deploy CI/CD signing workflow

ALGORITHM="Dilithium2"
ARTIFACT_HASH="sha256:..."
SIGNING_KEY="ci-signing-keys"

# Create/update artifact hash secret
kubectl create secret generic artifact-hash \
  --from-literal=digest="$ARTIFACT_HASH" \
  --dry-run=client -o yaml | kubectl apply -f -

# Sign the artifact
kubectl apply -f - <<EOF
apiVersion: qubesec.io/v1
kind: QuantumSignMessage
metadata:
  name: ci-sign-$(date +%s)
spec:
  algorithm: $ALGORITHM
  privateKeyRef:
    name: $SIGNING_KEY
  messageRef:
    name: artifact-hash
  outputSecretName: ci-artifact-signature
  messageKey: digest
EOF

# Wait for signing to complete
kubectl wait --for=condition=ready quantumsignmessage \
  -l ci-artifact-hash=$(echo -n "$ARTIFACT_HASH" | sha256sum | cut -d' ' -f1) \
  --timeout=60s

# Extract signature for artifact
SIGNATURE=$(kubectl get secret ci-artifact-signature \
  -o jsonpath='{.data.signature}' | base64 -d)

echo "Artifact signed with signature: $SIGNATURE"
```

---

## Useful kubectl Commands

```bash
# List all signing keys
kubectl get quantumsignaturekeypairs

# List all signing operations
kubectl get quantumsignmessages

# List all verification operations
kubectl get quantumverifysignatures

# Watch signing status
kubectl get quantumsignmessage -w

# View detailed status
kubectl describe quantumsignaturekeypair <name>

# Get YAML output
kubectl get quantumsignmessage <name> -o yaml

# Filter by algorithm
kubectl get quantumsignaturekeypairs \
  -o jsonpath='{range .items[?(@.spec.algorithm=="Dilithium2")]}{.metadata.name}{"\n"}{end}'

# Get all failed operations
kubectl get quantumsignmessages \
  -o jsonpath='{range .items[?(@.status.status=="Failed")]}{.metadata.name}{"\n"}{end}'
```

---

## Troubleshooting

### Signing Fails
```bash
# 1. Check if keypair exists and is successful
kubectl describe quantumsignaturekeypair my-signing-keys

# 2. Check if message secret exists
kubectl get secret my-message

# 3. Check signing pod logs
kubectl logs -l app=qubesec -c controller

# 4. Check detailed error
kubectl get quantumsignmessage sign-my-message -o jsonpath='{.status.error}'
```

### Verification Fails
```bash
# 1. Check if public key is available
kubectl get secret my-sig-secret -o jsonpath='{.data.public-key}'

# 2. Verify algorithm matches
kubectl get quantumsignmessage sign-my-message -o jsonpath='{.spec.algorithm}'
kubectl get quantumverifysignature verify-my-signature -o jsonpath='{.spec.algorithm}'

# 3. Check signature integrity
kubectl get secret my-signature-output -o jsonpath='{.data.signature}'
```

---

## Performance Considerations

- Signatures are created asynchronously
- Status is updated once signing completes
- For multiple signatures, create parallel QuantumSignMessage CRs
- Verification is read-only and can be run multiple times

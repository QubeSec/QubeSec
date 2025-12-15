# QubeSec Setup Guide

This document contains all installation, configuration, and operational commands for QubeSec.

## Quick Start

### Install QubeSec Operator (One Command)

```bash
kubectl apply -f https://raw.githubusercontent.com/QubeSec/QubeSec/main/dist/install.yaml
```

This single command installs:
- All Custom Resource Definitions (CRDs)
- QubeSec operator deployment
- RBAC configurations
- Namespace and services

### Verify Installation

```bash
kubectl get pods -n qubesec-system
kubectl api-resources | grep qubesec
```

### Create Your First Quantum Resources

```bash
kubectl apply -f https://raw.githubusercontent.com/QubeSec/QubeSec/main/config/samples/_v1_quantumrandomnumber.yaml
kubectl get qrn
```

---

## Environment Setup

### Using Ansible (Recommended)

The Ansible playbook automates all environment setup including liboqs, OpenSSL with oqs-provider, and Go bindings:

```bash
cd ansible
ansible-playbook setup-liboqs.yml
```

### Manual Setup

#### Prerequisites

```bash
sudo apt install make build-essential git cmake libssl-dev ninja-build pkg-config python3
```

#### Clone Repositories

```bash
git clone --depth 1 --branch 0.15.0 https://github.com/open-quantum-safe/liboqs
git clone --depth 1 --branch 0.12.0 https://github.com/open-quantum-safe/liboqs-go
```

#### Install liboqs

```bash
cmake -S liboqs -B liboqs/build -DCMAKE_PREFIX_PATH=/opt/liboqs -DBUILD_SHARED_LIBS=ON
cmake --build liboqs/build --parallel 4
sudo cmake --build liboqs/build --target install
```

#### Configure Environment Variables

Add to `~/.bashrc`:

```bash
export PKG_CONFIG_PATH=/opt/liboqs/lib/pkgconfig:/opt/liboqs-go/.config
export LD_LIBRARY_PATH=/opt/liboqs/lib:/opt/openssl/lib:/opt/openssl/lib64:${LD_LIBRARY_PATH}
export PATH=/opt/openssl/bin:${PATH}
```

## Kubernetes Operations

### Generate Manifests

```bash
make generate
make manifests
```

### Install CRDs

```bash
make install    # Install CRDs into cluster
make uninstall  # Remove CRDs from cluster
```

### Run Locally (Development)

```bash
make run
```

### Deploy to Cluster

```bash
# Quick Install (One-liner)
kubectl apply -f https://raw.githubusercontent.com/QubeSec/QubeSec/main/dist/install.yaml

# Or build and deploy from source
make deploy     # Deploy operator to cluster
make undeploy   # Remove operator from cluster
```

### Uninstall from Cluster

```bash
kubectl delete -f https://raw.githubusercontent.com/QubeSec/QubeSec/main/dist/install.yaml
```

### Apply Sample Resources

```bash
# Create sample resources
kubectl apply -k config/samples/

# Verify resource creation
kubectl get qkkp,qes,qds,qdk,qskp,qc,qrn,qsm,qvs

# View created secrets
kubectl get secrets

# Clean up samples
kubectl delete -k config/samples/
```

### Quantum Key Encapsulation & Derivation Workflow

The complete quantum-safe key exchange workflow uses three components working together:

#### Step 1: Create a KEM Key Pair
```bash
kubectl apply -f config/samples/_v1_quantumkemkeypair.yaml
kubectl get qkkp quantumkemkeypair-sample
```

This generates a Kyber (ML-KEM-1024) public/private keypair stored in a Secret.

#### Step 2: Encapsulate to Generate Shared Secret
```bash
kubectl apply -f config/samples/_v1_quantumencapsulatesecret.yaml
kubectl get qes quantumencapsulatesecret-sample
```

This uses the public key to encapsulate and produces:
- A **shared secret** stored in Kubernetes Secret `quantumencapsulatesecret-sample-sharedsecret`
- A **ciphertext** stored in the status field (hex-encoded)

#### Step 3: Retrieve and Use the Ciphertext
```bash
# Get the ciphertext from encapsulation for decapsulation
CIPHERTEXT=$(kubectl get qes quantumencapsulatesecret-sample -o jsonpath='{.status.ciphertext}')
echo "Ciphertext: $CIPHERTEXT"

# View the encapsulated shared secret
kubectl get secret encapsulated-shared-secret -o jsonpath='{.data.shared-secret}' | base64 -d | xxd -p
```

#### Step 4: Decapsulate Using Private Key and Ciphertext
```bash
# Update the decapsulate sample with the current ciphertext
kubectl patch -p '{"spec":{"ciphertext":"'$CIPHERTEXT'"}}' \
  --type merge qds quantumdecapsulatesecret-sample

# Or apply the pre-configured sample (ciphertext must be current)
kubectl apply -f config/samples/_v1_quantumdecapsulatesecret.yaml

# Verify the decapsulation
kubectl get qds quantumdecapsulatesecret-sample -o jsonpath='{.status.status}'
```

The controller will:
1. Retrieve the private key from the referenced QuantumKEMKeyPair
2. Decode the hex ciphertext from the spec
3. Perform KEM decapsulation to recover the shared secret
4. Store the recovered secret in Kubernetes Secret `quantumdecapsulatesecret-sample-sharedsecret`

#### Step 5: Verify Encapsulation/Decapsulation are Correct
```bash
# Get shared secret from encapsulation
ENCAP_SECRET=$(kubectl get secret quantumencapsulatesecret-sample-sharedsecret -o jsonpath='{.data.shared-secret}' | base64 -d | xxd -p)

# Get shared secret from decapsulation
DECAP_SECRET=$(kubectl get secret quantumdecapsulatesecret-sample-sharedsecret -o jsonpath='{.data.shared-secret}' | base64 -d | xxd -p)

# Compare them (must be identical for correct implementation)
if [ "$ENCAP_SECRET" = "$DECAP_SECRET" ]; then
  echo "✓ Secrets match! Decapsulation successful."
else
  echo "✗ Secrets don't match. Check ciphertext is current."
fi
```

#### Step 6: Derive Keys from Either Source
```bash
# Option A: Derive from encapsulated secret
kubectl apply -f config/samples/_v1_quantumderivedkey-from-encapsulated.yaml
kubectl get qdk quantumderivedkey-from-encapsulated

# Option B: Derive from decapsulated secret
kubectl apply -f config/samples/_v1_quantumderivedkey-from-decapsulated.yaml
kubectl get qdk quantumderivedkey-from-decapsulated
```

Both derived keys will be identical (same fingerprint) if the shared secrets match:
```bash
kubectl get qdk -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.fingerprint}{"\n"}{end}'
```

Expected output shows both resources with the same fingerprint:
```
quantumderivedkey-from-decapsulated    d1c312b81f
quantumderivedkey-from-encapsulated    d1c312b81f
```

## Docker Operations

### Build Consolidated Installer

Generate a single `install.yaml` file with all CRDs and operator resources:

```bash
export IMG=qubesec/qubesec:main
make build-installer
# Output: dist/install.yaml
```

### Build and Push Image

```bash
export IMG=qubesec/qubesec:v0.1.22
make docker-build docker-push
```

### Load Image to Minikube

```bash
make docker-build
minikube image load qubesec/qubesec:v0.1.22
```

## Key Storage and Retrieval

All cryptographic keys are stored in Kubernetes Secrets as **hex-encoded** binary data using the `Data` field.

### View Secret Keys

```bash
kubectl get secret <secret-name> -o jsonpath='{.data.<key>}' | base64 -d
```

### Inspect Quantum Certificates

```bash
# For post-quantum certificates with oqs-provider
kubectl get secret quantumcertificate-sample-cert \
  -o jsonpath='{.data.tls\.crt}' | \
  base64 -d | openssl x509 -text -noout
```

### Retrieve Public Key from QuantumKEMKeyPair

```bash
kubectl get secret quantumkemkeypair-sample-keypair \
  -o jsonpath='{.data.public-key}' | base64 -d
```

### Retrieve Public Key from QuantumSignatureKeyPair

```bash
kubectl get secret quantumsignaturekeypair-sample-keypair \
  -o jsonpath='{.data.public-key}' | base64 -d
```

## Cryptographic Architecture

### Workflow Chain

QubeSec implements a complete post-quantum cryptographic workflow:

```
QuantumKEMKeyPair (Kyber keypair)
  ↓
QuantumEncapsulateSecret (public key → encapsulation)
  ├─ Output: Shared Secret + Ciphertext
  └─ Can be recovered via:
      └─ QuantumDecapsulateSecret (private key + ciphertext → same shared secret)
        ↓
QuantumDerivedKey (shared secret → AES-256 key via HKDF-SHA256)
  ├─ Fingerprint: SHA256 hash of derived key
  └─ Can use either encapsulated or decapsulated secret as source
```

### Key Properties

- **Deterministic**: Same inputs always produce same outputs
- **Symmetric Verification**: If encapsulated and decapsulated secrets match, both derived keys will be identical
- **HKDF-SHA256 Derivation**: Uses optional salt/info parameters for additional entropy
- **Fingerprinting**: SHA256 hash of derived key for integrity verification
- **Binary Storage**: Keys stored as raw binary data in Kubernetes Secrets

## Custom Resource Abbreviations

```
qc   = QuantumCertificate
qdk  = QuantumDerivedKey
qds  = QuantumDecapsulateSecret
qes  = QuantumEncapsulateSecret
qkkp = QuantumKEMKeyPair
qrn  = QuantumRandomNumber
qskp = QuantumSignatureKeyPair
```

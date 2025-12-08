# QubeSec Setup Guide

This document contains all installation, configuration, and operational commands for QubeSec.

## Quick Start Commands

### Initialize the Project with Kubebuilder

```bash
kubebuilder init \
  --domain qubesec.io \
  --repo github.com/QubeSec/QubeSec
```

### Create APIs

```bash
# QuantumRandomNumber
kubebuilder create api \
  --version v1 \
  --kind QuantumRandomNumber \
  --resource \
  --controller

kubebuilder create webhook \
  --version v1 \
  --kind QuantumRandomNumber \
  --defaulting \
  --programmatic-validation

# QuantumKEMKeyPair
kubebuilder create api \
  --version v1 \
  --kind QuantumKEMKeyPair \
  --resource \
  --controller

# QuantumSignatureKeyPair
kubebuilder create api \
  --version v1 \
  --kind QuantumSignatureKeyPair \
  --resource \
  --controller

# QuantumCertificate
kubebuilder create api \
  --version v1 \
  --kind QuantumCertificate \
  --resource \
  --controller

# Future APIs (In Discussions)
kubebuilder create api \
  --version v1 \
  --kind QuantumDigitalSignature \
  --resource \
  --controller

kubebuilder create api \
  --version v1 \
  --kind KeyRequest \
  --resource \
  --controller
```

## Environment Setup

### Using Ansible (Recommended)

The Ansible playbook automates all environment setup including liboqs, OpenSSL with oqs-provider, and Go bindings:

```bash
cd ansible
ansible-playbook setup-liboqs.yml -i hosts.yml
```

See [Ansible Architecture](ansible/ARCHITECTURE.md) for detailed role descriptions.

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
# Build and deploy
make deploy     # Deploy operator to cluster
make undeploy   # Remove operator from cluster
```

### Apply Sample Resources

```bash
# Create sample resources
kubectl apply -k config/samples/

# Verify resource creation
kubectl get qkkp,qss,qdk,qskp,qc,qrn

# Clean up samples
kubectl delete -k config/samples/
```

## Docker Operations

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
kubectl get secret quantumcertificate-sample-tls \
  -o jsonpath='{.data.tls\.crt}' | \
  base64 -d | openssl x509 -text -noout
```

### Retrieve Public Key from QuantumKEMKeyPair

```bash
kubectl get secret quantumkemkeypair-sample-keys \
  -o jsonpath='{.data.public-key}' | base64 -d
```

### Retrieve Public Key from QuantumSignatureKeyPair

```bash
kubectl get secret quantumsignaturekeypair-sample-keys \
  -o jsonpath='{.data.public-key}' | base64 -d
```

## Custom Resource Abbreviations

```
qc   = QuantumCertificate
qdk  = QuantumDerivedKey
qkkp = QuantumKEMKeyPair
qrn  = QuantumRandomNumber
qss  = QuantumSharedSecret
qskp = QuantumSignatureKeyPair
```

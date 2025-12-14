# QubeSec

<div align="center">
  <img src="https://raw.githubusercontent.com/QubeSec/QubeSec/refs/heads/main/assets/qubesec.png" alt="QubeSec" width="400">
</div>

A Kubernetes operator for post-quantum cryptography providing custom resource definitions (CRDs) and controllers for quantum-safe key generation, key encapsulation, key derivation, and certificate management.

## Overview

QubeSec leverages [liboqs](https://github.com/open-quantum-safe/liboqs) and [OpenSSL with oqs-provider](https://github.com/open-quantum-safe/oqs-provider) to integrate post-quantum cryptographic algorithms into Kubernetes. All cryptographic operations are automated through custom controllers that orchestrate the NIST-standardized quantum-safe algorithms (Kyber, Dilithium, etc.).

## Key Features

- **Quantum-Safe Key Generation**: Generate Kyber KEM keypairs and Dilithium signature keypairs
- **Key Encapsulation**: Derive shared secrets using KEM encapsulation from public keys
- **Key Decapsulation**: Recover shared secrets using KEM decapsulation with private key and ciphertext
- **Key Derivation**: Generate AES-256 keys from shared secrets using HKDF-SHA256
- **Quantum Certificates**: Create X.509 certificates with post-quantum algorithms
- **Secure Secret Storage**: All keys stored as raw binary data in Kubernetes Secrets
- **Key Fingerprinting**: SHA256 fingerprints generated for shared secrets (encap/decap) and derived keys for verification and audit
- **Ciphertext Bridging**: Decapsulation can pull ciphertext directly from a referenced QuantumEncapsulateSecret status (or via inline spec.ciphertext)
- **Automated Workflows**: Chainable controllers (KEM → Shared Secret → Derived Key)

## Supported Algorithms

- **Key Encapsulation**: Kyber (NIST-standardized post-quantum KEM)
- **Signature**: Dilithium (NIST-standardized post-quantum signature algorithm)
- **Random Generation**: Go's `crypto/rand` with cryptographically secure randomness

## Getting Started

### Quick Install (Kubernetes)

Install QubeSec operator and all CRDs with a single command:

```bash
kubectl apply -f https://raw.githubusercontent.com/QubeSec/QubeSec/main/dist/install.yaml
```

Verify installation:

```bash
kubectl get pods -n qubesec-system
kubectl get crd | grep qubesec
```

Create your first quantum resource:

```bash
kubectl apply -f https://raw.githubusercontent.com/QubeSec/QubeSec/main/config/samples/_v1_quantumrandomnumber.yaml
kubectl get qrn
```

### Development Setup with Ansible

```bash
cd ansible
ansible-playbook setup-liboqs.yml
```

This automates installation of liboqs, OpenSSL with oqs-provider, and Go bindings to `/opt/`.

### Manual Setup

For step-by-step instructions, environment variable configuration, and command reference, see [SETUP.md](SETUP.md).

## Infrastructure Setup

## Development

See [SETUP.md](SETUP.md) for:
- Kubebuilder initialization and API generation
- CRD installation and testing
- Local development and debugging
- Docker image building and deployment

## Documentation

- [SETUP.md](SETUP.md) - Complete installation and operation guide
- [api/v1/](api/v1/) - Custom Resource Definitions

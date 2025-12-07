# QubeSec

<div align="center">
  <img src="https://raw.githubusercontent.com/QubeSec/QubeSec/refs/heads/main/assets/qubesec.png" alt="QubeSec" width="400">
</div>

A Kubernetes operator for post-quantum cryptography providing custom resource definitions (CRDs) and controllers for quantum-safe key generation, key encapsulation, key derivation, and certificate management.

## Overview

QubeSec leverages [liboqs](https://github.com/open-quantum-safe/liboqs) and [OpenSSL with oqs-provider](https://github.com/open-quantum-safe/oqs-provider) to integrate post-quantum cryptographic algorithms into Kubernetes. All cryptographic operations are automated through custom controllers that orchestrate the NIST-standardized quantum-safe algorithms (Kyber, Dilithium, etc.).

### Key Features

- **Quantum-Safe Key Generation**: Generate Kyber KEM keypairs and Dilithium signature keypairs
- **Key Encapsulation**: Derive shared secrets using KEM encapsulation from public keys
- **Key Derivation**: Generate AES-256 keys from shared secrets using HKDF-SHA256
- **Quantum Certificates**: Create X.509 certificates with post-quantum algorithms
- **Secure Secret Storage**: All keys stored as raw binary data in Kubernetes Secrets
- **Key Fingerprinting**: SHA256 fingerprints generated for derived keys for verification and audit
- **Automated Workflows**: Chainable controllers (KEM → Shared Secret → Derived Key)

### Supported Algorithms

- **Key Encapsulation**: Kyber (NIST-standardized post-quantum KEM)
- **Signature**: Dilithium (NIST-standardized post-quantum signature algorithm)
- **Random Generation**: Go's `crypto/rand` with cryptographically secure randomness

## Custom Resources

| Abbreviation | Resource | Purpose |
|---|---|---|
| `qrn` | QuantumRandomNumber | Generate cryptographically secure random bytes |
| `qkkp` | QuantumKEMKeyPair | Generate Kyber KEM public/private keypairs |
| `qss` | QuantumSharedSecret | Derive shared secrets via KEM encapsulation |
| `qdk` | QuantumDerivedKey | Derive AES-256 keys from shared secrets using HKDF |
| `qskp` | QuantumSignatureKeyPair | Generate Dilithium signature keypairs |
| `qc` | QuantumCertificate | Create X.509 certificates with quantum algorithms |

## Getting Started

### Quick Setup with Ansible

```bash
cd ansible
ansible-playbook setup-liboqs.yml -i hosts.yml
```

This automates installation of liboqs, OpenSSL with oqs-provider, and Go bindings to `/opt/`.

### Manual Setup

For step-by-step instructions, environment variable configuration, and command reference, see [SETUP.md](SETUP.md).

## Key Storage Format

All cryptographic keys (keypairs, certificates, derived keys, shared secrets, random numbers) are stored in Kubernetes Secrets in raw binary data format. This ensures secure and efficient storage.

### QuantumDerivedKey Fingerprint

When deriving keys using QuantumDerivedKey, a **fingerprint** is automatically generated for the derived key. The fingerprint is a SHA256 hash of the derived key and serves the following purposes:

- **Verification**: Verify key integrity and authenticity without exposing the full key
- **Identification**: Uniquely identify derived keys for audit and compliance purposes
- **Status Tracking**: Included in the QuantumDerivedKey status for transparency

The fingerprint is stored in **hex-encoded format** for human readability and is available in both the Kubernetes Secret and the status field of the QuantumDerivedKey resource.

For retrieval and inspection examples, see [SETUP.md - Key Storage and Retrieval](SETUP.md#key-storage-and-retrieval).

## Architecture

### Workflow Example: Kyber Key Encapsulation and Derivation

```
1. Create QuantumKEMKeyPair
   ↓
2. Create QuantumSharedSecret (references KEM keypair, derives via encapsulation)
   ↓
3. Create QuantumDerivedKey (references shared secret, derives AES-256 key via HKDF)
```

All intermediate results are stored in Kubernetes Secrets for consumption by other workloads.

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

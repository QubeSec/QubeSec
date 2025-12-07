# QubeSec

A Kubernetes operator for post-quantum cryptography providing custom resource definitions (CRDs) and controllers for quantum-safe key generation, key encapsulation, key derivation, and certificate management.

## Overview

QubeSec leverages [liboqs](https://github.com/open-quantum-safe/liboqs) and [OpenSSL with oqs-provider](https://github.com/open-quantum-safe/oqs-provider) to integrate post-quantum cryptographic algorithms into Kubernetes. All cryptographic operations are automated through custom controllers that orchestrate the NIST-standardized quantum-safe algorithms (Kyber, Dilithium, etc.).

### Key Features

- **Quantum-Safe Key Generation**: Generate Kyber KEM keypairs and Dilithium signature keypairs
- **Key Encapsulation**: Derive shared secrets using KEM encapsulation from public keys
- **Key Derivation**: Generate AES-256 keys from shared secrets using HKDF-SHA256
- **Quantum Certificates**: Create X.509 certificates with post-quantum algorithms
- **Secure Secret Storage**: All keys stored as hex-encoded data in Kubernetes Secrets
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

All cryptographic keys (keypairs, certificates, derived keys, shared secrets, random numbers) are stored in Kubernetes Secrets as **hex-encoded** binary data. This ensures consistent and secure binary storage.

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

For detailed information on role structure, dependencies, and path configuration, see [ansible/ARCHITECTURE.md](ansible/ARCHITECTURE.md).

## Development

See [SETUP.md](SETUP.md) for:
- Kubebuilder initialization and API generation
- CRD installation and testing
- Local development and debugging
- Docker image building and deployment

## Documentation

- [SETUP.md](SETUP.md) - Complete installation and operation guide
- [ansible/ARCHITECTURE.md](ansible/ARCHITECTURE.md) - Ansible roles and infrastructure design
- [api/v1/](api/v1/) - Custom Resource Definitions

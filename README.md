# QubeSec

<div align="center">
  <img src="https://raw.githubusercontent.com/QubeSec/QubeSec/refs/heads/main/assets/qubesec.png" alt="QubeSec" width="400">
</div>

A Kubernetes operator for post-quantum cryptography providing custom resource definitions (CRDs) and controllers for quantum-safe key generation, key encapsulation, key derivation, and certificate management.

## Overview

QubeSec leverages [liboqs](https://github.com/open-quantum-safe/liboqs) and [OpenSSL with oqs-provider](https://github.com/open-quantum-safe/oqs-provider) to integrate post-quantum cryptographic algorithms into Kubernetes. All cryptographic operations are automated through custom controllers that orchestrate the NIST-standardized quantum-safe algorithms (Kyber, Dilithium, etc.).

## Key Features

- **Quantum-Safe Key Generation**: Generate Kyber KEM keypairs and Dilithium/Falcon/SPHINCS+ signature keypairs
- **Key Encapsulation**: Derive shared secrets using KEM encapsulation from public keys
- **Key Decapsulation**: Recover shared secrets using KEM decapsulation with private key and ciphertext
- **Key Derivation**: Generate AES-256 keys from shared secrets using HKDF-SHA256
- **Quantum Signatures**: Sign messages and verify signatures with post-quantum algorithms (ML-DSA, SLH-DSA)
- **Quantum Certificates**: Create X.509 certificates with post-quantum algorithms
- **Random Number Generation**: Generate cryptographically secure random bytes
- **Secure Secret Storage**: All keys stored as raw binary data in Kubernetes Secrets
- **Key Fingerprinting**: SHA256 fingerprints for keys, messages, and secrets for verification without exposing material
- **Ciphertext Bridging**: Decapsulation can pull ciphertext directly from a referenced QuantumEncapsulateSecret status
- **Automated Workflows**: Chainable controllers (KEM → Shared Secret → Derived Key)

## Supported Algorithms

- **Key Encapsulation**: Kyber512/768/1024 (ML-KEM - NIST-standardized post-quantum KEM)
- **Digital Signatures**: Dilithium2/3/5 (ML-DSA), Falcon512/1024, SPHINCS+-SHA2 (NIST post-quantum signatures)
- **Random Generation**: Cryptographically secure random number generation via `crypto/rand`

## Documentation

- [SETUP.md](docs/SETUP.md) - Complete installation and operation guide
- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - System architecture and design patterns
- [OPENSSL.md](docs/OPENSSL.md) - OpenSSL command reference for quantum-safe operations
- [api/v1/](api/v1/) - Custom Resource Definitions

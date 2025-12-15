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

## Why Post-Quantum Cryptography Now?

Governments and cybersecurity agencies worldwide have issued official guidance directing organizations to begin migrating to post-quantum cryptography. Below are key authoritative documents:

### Government Directives & Agency Guidance

1. **[White House / OMB — Memorandum M-23-02](https://www.whitehouse.gov/wp-content/uploads/2022/11/M-23-02-M-Memo-on-Migrating-to-Post-Quantum-Cryptography.pdf) (November 2022)**
   - Formal U.S. government memo instructing federal agencies to prepare and begin migration planning to post-quantum cryptography

2. **[CISA / NSA / NIST — Quantum Readiness Resource](https://www.cisa.gov/news-events/news/cisa-nsa-and-nist-publish-new-resource-migrating-post-quantum-cryptography) (August 2023)**
   - Joint playbook recommending organizations start now with inventory, vendor engagement, and roadmap planning

3. **[NIST — IR 8547: Transition to Post-Quantum Cryptography Standards](https://csrc.nist.gov/pubs/ir/8547/ipd) (2024)**
   - Technical guidance on transitioning from vulnerable algorithms to NIST-standardized PQC algorithms

4. **[UK NCSC — Next Steps in Preparing for Post-Quantum Cryptography](https://www.ncsc.gov.uk/whitepaper/next-steps-preparing-for-post-quantum-cryptography) (March 2025)**
   - Explicit timelines and migration steps for UK organizations with concrete deadlines

5. **[ENISA — Post-Quantum Cryptography Reports](https://www.enisa.europa.eu/publications/post-quantum-cryptography-current-state-and-quantum-mitigation) (2024)**
   - EU-level guidance on migration planning and readiness for member states

6. **[DoD / NSA — CNSA Suite 2.0](https://media.defense.gov/2022/Sep/07/2003071836/-1/-1/0/CSI_CNSA_2.0_FAQ_.PDF) (September 2022)**
   - Defense-grade requirements for quantum-resistant algorithms in national security systems

7. **[India TEC — Migration to Post-Quantum Cryptography](https://www.tec.gov.in/pdf/TR/Final%20technical%20report%20on%20migration%20to%20PQC%2028-03-25.pdf) (March 2025)**
   - National technical guidance for critical infrastructure quantum-safe measures

## Documentation

- [SETUP.md](docs/SETUP.md) - Complete installation and operation guide
- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - System architecture and design patterns
- [OPENSSL.md](docs/OPENSSL.md) - OpenSSL command reference for quantum-safe operations
- [api/v1/](api/v1/) - Custom Resource Definitions

# QubeSec Installation

## Quick Install

Install QubeSec operator and all CRDs with a single command:

```bash
kubectl apply -f https://raw.githubusercontent.com/QubeSec/QubeSec/main/dist/install.yaml
```

Or if you have the file locally:

```bash
kubectl apply -f install.yaml
```

## What Gets Installed

- **Namespace**: `qubesec-system`
- **CRDs**: All 6 quantum resource definitions
- **Operator**: QubeSec controller manager
- **RBAC**: Service accounts, roles, and bindings
- **Services**: Metrics endpoint

## Verify Installation

```bash
kubectl get pods -n qubesec-system
kubectl get crd | grep qubesec
```

## Create Sample Resources

```bash
kubectl apply -f https://raw.githubusercontent.com/QubeSec/QubeSec/main/config/samples/_v1_quantumrandomnumber.yaml
kubectl get qrn
```

## Uninstall

```bash
kubectl delete -f https://raw.githubusercontent.com/QubeSec/QubeSec/main/dist/install.yaml
```

## Building from Source

To regenerate this file:

```bash
export IMG=qubesec/qubesec:main
make build-installer
```

# QubeSec

## Key Storage Format

All cryptographic keys (PEM-encoded keypairs, certificates, derived keys, shared secrets, random numbers) are stored in Kubernetes Secrets as **hex-encoded** binary data using the `Data` field (not `StringData`). This ensures consistent and secure binary storage.

To decode and inspect:
```bash
# View first 100 hex characters
kubectl get secret <secret-name> -o jsonpath='{.data.<key>}' | base64 -d

# Convert hex to ASCII (for PEM keys)
kubectl get secret <secret-name> -o jsonpath='{.data.<key>}' | base64 -d | xxd -r -p
```

---

Initialize the project with kubebuilder:
```bash
kubebuilder init \
  --domain qubesec.io \
  --repo github.com/QubeSec/QubeSec
```

Create a new API:
```bash
# Implemented:
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

kubebuilder create api \
  --version v1 \
  --kind QuantumKEMKeyPair \
  --resource \
  --controller

kubebuilder create api \
  --version v1 \
  --kind QuantumSignatureKeyPair \
  --resource \
  --controller

kubebuilder create api \
  --version v1 \
  --kind QuantumCertificate \
  --resource \
  --controller

# In Discussions:
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

Generate the manifests:
```bash
make generate
make manifests
```

Install CRDs into the Kubernetes cluster using kubectl apply:
```bash
make install
make uninstall
```

Regenerate code and run against the Kubernetes cluster configured by `~/.kube/config`:
```bash
export ENABLE_WEBHOOKS=false
make run
```

Apply samples for testing operator:
```bash
kubectl apply -k config/samples/
kubectl delete -k config/samples/
```

Get the custom resources:
```bash
kubectl get qc,qdk,qkkp,qrn,qss,qskp
```

Export the docker image:
```bash
export IMG=qubesec/qubesec:v0.1.22
```

Build the docker image:
```bash
make docker-build docker-push
```

Install cert-manager:
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.15.3/cert-manager.yaml
```

Load the docker image into minikube:
```bash
make docker-build && minikube image load qubesec/qubesec:v0.1.22
```

Create a deployment:
```bash
make deploy
make undeploy
```

Generate certificates for serving the webhooks locally:
```bash
mkdir -p /tmp/k8s-webhook-server/serving-certs/
cd /tmp/k8s-webhook-server/serving-certs/
openssl req -newkey rsa:2048 -nodes -keyout tls.key -x509 -days 365 -out tls.crt
```

Check the certificate info (note: keys are now stored as hex):
```bash
# For post-quantum certificates with oqs-provider:
kubectl get secret quantumcertificate-sample-tls \
  -o jsonpath='{.data.tls\.crt}' | \
  base64 -d | xxd -r -p | openssl x509 -text -noout

# Or view the raw hex:
kubectl get secret quantumcertificate-sample-tls \
  -o jsonpath='{.data.tls\.crt}' | base64 -d | head -c 100
```

**Note**: Quantum certificate generation requires OpenSSL with oqs-provider loaded. See [Ansible Role Setup](#setting-up-with-ansible) for building OpenSSL with oqs-provider support.

---

### Enviourment Setup

Install prerequsit:
```bash
sudo apt install make build-essential git cmake libssl-dev
```

Clone liboqs and liboqs-go:
```bash
git clone --depth 1 --branch 0.15.0 https://github.com/open-quantum-safe/liboqs
git clone --depth 1 --branch 0.12.0 https://github.com/open-quantum-safe/liboqs-go
```

Install liboqs:
```bash
cmake -S liboqs -B liboqs/build -DBUILD_SHARED_LIBS=ON
cmake --build liboqs/build --parallel 4
sudo cmake --build liboqs/build --target install
```

Set environment variables: `vim ~/.bashrc`
```bash
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/home/ubuntu/liboqs-go/.config
export LD_LIBRARY_PATH=/usr/local/lib
```

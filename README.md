# QubeSec

Initialize the project with kubebuilder:
```bash
kubebuilder init \
  --domain qubesec.io \
  --repo github.com/QubeSec/QubeSec
```

Create a new API:
```bash
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
  --kind QuantumDigitalSignature \
  --resource \
  --controller

kubebuilder create api \
  --version v1 \
  --kind KeyRequest \
  --resource \
  --controller

kubebuilder create api \
  --version v1 \
  --kind Certificate \
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

Create a KeyRequest Custom Resource:
```bash
kubectl apply -k config/samples/
kubectl delete -k config/samples/
```

Export the docker image:
```bash
export IMG=qubesec/qubesec:v0.1.1
```

Build the docker image:
```bash
make docker-build docker-push
```

Install cert-manager:
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.2/cert-manager.yaml
```

Load the docker image into minikube:
```bash
minikube image load qubesec/qubesec:v0.1.2
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

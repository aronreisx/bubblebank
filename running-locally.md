# Running bubblebank locally

## Prerequisites

- Docker
- Kind
- Kubectl
- Helm

## Instructions

### 1. Create a Kind cluster with ingress support

```sh
cat <<EOF | kind create cluster --name local --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF
```

### 2. Install NGINX Ingress Controller

```sh
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

Wait for the ingress controller to be ready:

```sh
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```

### 3. Load your Docker image into Kind

After building your Docker image, load it into Kind:

```sh
kind load docker-image aronreis/bubblebank:latest --name local
```

### 4. Install the application using Helm

```sh
helm install --namespace bubblebank --create-namespace bubblebank ./helm/bubblebank
helm install --namespace argocd --create-namespace argocd ./helm/bubblebank
```

### 5. Access the application

Open your browser and navigate to:

```
http://bubblebank.api.local
```

## Cleanup

```sh
kind delete cluster --name local
```

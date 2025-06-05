
set -e

echo "=== BubbleBank Local Environment Setup ==="
echo ""

echo "Checking prerequisites..."
for cmd in docker kind kubectl helm; do
    if ! command -v $cmd &> /dev/null; then
        echo "Error: $cmd is not installed. Please install it before continuing."
        exit 1
    fi
done

echo "All prerequisites are installed!"
echo ""

echo "Step 1: Creating Kind cluster with ingress support..."
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

echo "Kind cluster created successfully!"
echo ""

echo "Step 2: Installing NGINX Ingress Controller..."
helm dependency update infra/helm/nginx-ingress
helm install nginx-ingress infra/helm/nginx-ingress --create-namespace --namespace ingress-nginx

echo "Waiting for ingress controller to be ready..."

echo "Waiting for ingress-nginx-controller deployment to be available..."
kubectl wait --namespace ingress-nginx \
--for=condition=available deployment/ingress-nginx-controller \
--timeout=120s || {
    echo "Timed out waiting for ingress-nginx-controller deployment"
    echo "Checking deployment status..."
    kubectl get deployment -n ingress-nginx
    echo "Checking pods status..."
    kubectl get pods -n ingress-nginx
    exit 1
}

echo "Ensuring NGINX Ingress admission webhook is fully ready..."
ATTEMPTS=0
MAX_ATTEMPTS=24
WEBHOOK_READY=false
while [ $ATTEMPTS -lt $MAX_ATTEMPTS ]; do
    if kubectl get endpoints -n ingress-nginx ingress-nginx-controller-admission -o jsonpath='{.subsets[?(@.addresses)].addresses[0].ip}' > /dev/null 2>&1; then
        echo "NGINX Ingress admission webhook endpoints are available."
        sleep 5
        if kubectl get validatingwebhookconfiguration ingress-nginx-admission > /dev/null 2>&1; then
            echo "ValidatingWebhookConfiguration 'ingress-nginx-admission' found."
            WEBHOOK_READY=true
            break
        else
            echo "ValidatingWebhookConfiguration 'ingress-nginx-admission' not found yet. Retrying..."
        fi
    fi
    ATTEMPTS=$((ATTEMPTS + 1))
    echo "Waiting for NGINX Ingress admission webhook to be ready (attempt $ATTEMPTS/$MAX_ATTEMPTS)..."
    sleep 5
done

echo "Step 3: Loading Docker image into Kind..."
echo "Note: Make sure you have built the Docker image 'aronreis/bubblebank:latest' before running this script."
read -p "Continue with loading the image? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    kind load docker-image aronreis/bubblebank:latest --name local
    echo "Docker image loaded successfully!"
else
    echo "Skipping Docker image loading. Remember to load it manually before installing the application."
fi
echo ""

# argocd-autopilot repo bootstrap --recover --app "${GIT_REPO}/infra/argocd/bootstrap/argo-cd"

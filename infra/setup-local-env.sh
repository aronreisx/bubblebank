#!/bin/bash

set -e

echo "=== BubbleBank Local Environment Setup ==="
echo ""

# Check for required tools
echo "Checking prerequisites..."
for cmd in docker kind kubectl helm; do
    if ! command -v $cmd &> /dev/null; then
        echo "Error: $cmd is not installed. Please install it before continuing."
        exit 1
    fi
done

echo "All prerequisites are installed!"
echo ""

# Step 1: Create a Kind cluster with ingress support
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

# Step 2: Install NGINX Ingress Controller
echo "Step 2: Installing NGINX Ingress Controller..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

echo "Waiting for ingress controller to be ready..."

# Wait for the deployment to be available
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
MAX_ATTEMPTS=24 # Wait for up to 2 minutes (24 * 5s)
WEBHOOK_READY=false
while [ $ATTEMPTS -lt $MAX_ATTEMPTS ]; do
    # Check if the admission service has populated endpoints
    if kubectl get endpoints -n ingress-nginx ingress-nginx-controller-admission -o jsonpath='{.subsets[?(@.addresses)].addresses[0].ip}' > /dev/null 2>&1; then
        echo "NGINX Ingress admission webhook endpoints are available."
        # Add a very short grace period for the webhook to be fully operational after endpoints are up.
        sleep 5
        # Verify the ValidatingWebhookConfiguration is present
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

# if [ "$WEBHOOK_READY" = "false" ]; then
#     echo "Timed out waiting for NGINX Ingress admission webhook."
#     echo "Dumping details for ingress-nginx namespace:"
#     echo "--- Pods ---"
#     kubectl get pods -n ingress-nginx -o wide
#     echo "--- Services ---"
#     kubectl get svc -n ingress-nginx
#     echo "--- Endpoints ---"
#     kubectl get endpoints -n ingress-nginx
#     echo "--- Deployment: ingress-nginx-controller ---"
#     kubectl describe deployment -n ingress-nginx ingress-nginx-controller
#     echo "--- Service: ingress-nginx-controller-admission ---"
#     kubectl describe service -n ingress-nginx ingress-nginx-controller-admission
#     echo "--- Endpoints: ingress-nginx-controller-admission ---"
#     kubectl describe endpoints -n ingress-nginx ingress-nginx-controller-admission
#     echo "--- ValidatingWebhookConfiguration: ingress-nginx-admission ---"
#     kubectl describe validatingwebhookconfiguration ingress-nginx-admission
#     echo "--- Logs from ingress-nginx controller pods ---"
#     PODS=$(kubectl get pods -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx,app.kubernetes.io/component=controller -o jsonpath='{.items[*].metadata.name}' 2>/dev/null)
#     if [ -n "$PODS" ]; then
#         for POD_NAME in $PODS; do
#             echo "--- Logs for $POD_NAME (last 50 lines) ---"
#             kubectl logs -n ingress-nginx "$POD_NAME" --tail=50 || echo "Failed to get logs for $POD_NAME"
#         done
#     else
#         echo "No controller pods found to retrieve logs."
#     fi
#     exit 1
# fi

# echo "NGINX Ingress Controller installed successfully!"
# echo ""

# Step 3: Load Docker image into Kind
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

# Step 4: Apply ArgoCD CRDs
# echo "Step 4: Applying ArgoCD Custom Resource Definitions (CRDs)..."
# kubectl apply -k https://github.com/argoproj/argo-cd/manifests/crds\?ref\=stable
# if [ $? -ne 0 ]; then
#     echo "Error applying ArgoCD CRDs. Please check your network connection or the URL."
#     exit 1
# fi
# echo "ArgoCD CRDs applied successfully!"
# echo ""

# Step 5: Install the ArgoCD using Helm
echo "Step 5: Installing ArgoCD using Helm..."

# Adopt already created CRDs for Helm management
# This is to ensure Helm can manage CRDs that might have been applied by the kubectl apply -k command
# or existed from a previous installation.
# ARGOCD_NAMESPACE="argocd"
# ARGOCD_RELEASE_NAME="argocd" # This should match the release name used in the helm install command

# echo "Adopting existing ArgoCD CRDs for Helm management..."
# for crd in "applications.argoproj.io" "applicationsets.argoproj.io" "argocdextensions.argoproj.io" "appprojects.argoproj.io"; do
#     echo "Processing CRD: $crd"
#     # Check if the CRD exists before trying to label/annotate
#     if kubectl get crd "$crd" -o name > /dev/null 2>&1; then
#         # The following commands label and annotate the CRDs.
#         kubectl label --overwrite crd "$crd" app.kubernetes.io/managed-by=Helm
#         kubectl annotate --overwrite crd "$crd" meta.helm.sh/release-namespace="$ARGOCD_NAMESPACE"
#         kubectl annotate --overwrite crd "$crd" meta.helm.sh/release-name="$ARGOCD_RELEASE_NAME"
#     else
#         echo "CRD $crd not found, skipping adoption."
#     fi
# done
# echo "CRD adoption process complete."

helm install --namespace argocd --create-namespace argocd ./helm/argocd

echo "ArgoCD installed successfully!"
echo ""

# Step 6: Install the application using Helm
echo "Step 6: Installing the Bubblebank API using Helm..."
helm install --namespace bubblebank --create-namespace bubblebank ./helm/bubblebank

echo "Bubblebank API installed successfully!"
echo ""

# Step 7: Display access information
echo "=== Setup Complete! ==="
echo "You can now access the application at: http://bubblebank.api.local"
echo "You can now access the application at: http://bubblebank.argo.local"
echo ""
echo "You may need to add the following entry to your /etc/hosts file:"
echo "127.0.0.1 bubblebank.api.local"
echo "127.0.0.1 bubblebank.argo.local"
echo ""
echo "To clean up the environment, run: kind delete cluster --name local"

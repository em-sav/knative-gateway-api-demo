REPO_DIR="$(dirname "$(realpath "$0")")/.."

# Create local Cluster
kind create cluster --config="$REPO_DIR/cluster/kind/config.yaml" --name=cncf-cluster

# Install gateway API CRDs
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.3.0/standard-install.yaml

# Install and Configure Envoy Gateway
# helm upgrade --install eg oci://docker.io/envoyproxy/gateway-helm --set config.envoyGateway.provider.kubernetes.deploy.type=GatewayNamespace --version v1.5.0 -n envoy-gateway-system --create-namespace
# kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
# kubectl apply -f "$REPO_DIR/cluster/envoy-gateway/"

# Install Istio
helm repo add istio-official https://istio-release.storage.googleapis.com/charts
helm upgrade --install istio-base istio-official/base --version 1.27.1 -n istio-system --create-namespace
helm upgrade --install istiod istio-official/istiod --version 1.27.1 -n istio-system
kubectl wait --timeout=5m -n istio-system deployment/istiod --for=condition=Available

kubectl apply -f "$REPO_DIR/cluster/istio/"

# Install and Configure MetalLB
helm repo add metallb https://metallb.github.io/metallb
helm upgrade --install metallb metallb/metallb --version 0.15.2 -n metallb-system --create-namespace
kubectl wait --timeout=5m -n metallb-system deployment/metallb-controller --for=condition=Available
kubectl apply -f "$REPO_DIR/cluster/metallb/"

# Install and Configure Knative Gateway
kubectl apply -f "$REPO_DIR/cluster/gateway/"
kubectl wait --timeout=5m -n gateway gateway/knative-gateway --for=condition=Programmed

# Install and Configure Knative Operator
helm repo add knative-operator https://knative.github.io/operator
helm upgrade --install knative-operator knative-operator/knative-operator --version 1.18.0 -n knative-operator --create-namespace
kubectl wait --timeout=5m -n knative-operator deployment/knative-operator --for=condition=Available

kubectl apply -f "$REPO_DIR/cluster/knative/01-ns.yaml"
kubectl apply -f "$REPO_DIR/cluster/knative/02-knative-serving.yaml"
kubectl label namespace knative-serving istio-injection=enabled # We need to manually label the knative-serving namespace to be in the Istio dataplane
sleep 15
kubectl wait --timeout=5m -n knative-serving deployment/controller --for=condition=Available

kubectl apply -f "$REPO_DIR/cluster/knative/03-net-gateway-api.yaml"
kubectl wait --timeout=5m -n knative-serving deployment/net-gateway-api-controller --for=condition=Available

# Configure Workload Namespace for demo
kubectl apply -f "$REPO_DIR/cluster/workload/"

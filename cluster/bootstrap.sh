REPO_DIR="$(dirname "$(realpath "$0")")/.."

# Create local Cluster
docker network create --driver bridge --subnet=172.18.0.0/16 --gateway=172.18.0.1 kind 2>/dev/null || true
kind create cluster --config="$REPO_DIR/cluster/kind/config.yaml" --name=cncf-cluster

# Install gateway API CRDs
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.0/standard-install.yaml

# Install and Configure Envoy Gateway
helm upgrade --install eg oci://docker.io/envoyproxy/gateway-helm -f "$REPO_DIR/cluster/envoy-gateway/helm/values.yaml" --version v1.6.1 -n envoy-gateway-system --create-namespace
kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
kubectl apply -f "$REPO_DIR/cluster/envoy-gateway/"

# Install and Configure MetalLB
helm repo add metallb https://metallb.github.io/metallb
helm upgrade --install metallb metallb/metallb --version 0.15.2 -n metallb-system --create-namespace
kubectl wait --timeout=5m -n metallb-system deployment/metallb-controller --for=condition=Available
kubectl apply -f "$REPO_DIR/cluster/metallb/"

# Install and Configure Knative Gateway
kubectl apply -f "$REPO_DIR/cluster/gateway/"
kubectl wait --timeout=5m -n gateway gateway/public-gateway --for=condition=Programmed
kubectl wait --timeout=5m -n gateway gateway/knative-gateway --for=condition=Programmed

# Install and Configure Knative Operator
helm repo add knative-operator https://knative.github.io/operator
helm upgrade --install knative-operator knative-operator/knative-operator --version 1.20.0 -n knative-operator --create-namespace
kubectl wait --timeout=5m -n knative-operator deployment/knative-operator --for=condition=Available

kubectl apply -f "$REPO_DIR/cluster/knative/01-ns.yaml"
kubectl apply -f "$REPO_DIR/cluster/knative/02-knative-serving.yaml"
kubectl label namespace knative-serving istio-injection=enabled
sleep 15
kubectl wait --timeout=5m -n knative-serving deployment/controller --for=condition=Available

kubectl apply -f "$REPO_DIR/cluster/knative/03-net-gateway-api.yaml"
kubectl wait --timeout=5m -n knative-serving deployment/net-gateway-api-controller --for=condition=Available

# Configure Workload Namespace for demo
kubectl apply -f "$REPO_DIR/cluster/workload/"

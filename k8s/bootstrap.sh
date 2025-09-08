REPO_DIR="$(dirname "$(realpath "$0")")/.."

# Create local Cluster
kind create cluster --config="$REPO_DIR/k8s/kind/config.yaml" --name=cncf-cluster

# Install and Configure Envoy Gateway
helm upgrade --install eg oci://docker.io/envoyproxy/gateway-helm --set config.envoyGateway.provider.kubernetes.deploy.type=GatewayNamespace --version v1.5.0 -n envoy-gateway-system --create-namespace
kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
kubectl apply -f "$REPO_DIR/k8s/envoy-gateway/"

# Install and Configure MetalLB
helm repo add metallb https://metallb.github.io/metallb
helm upgrade --install metallb metallb/metallb --version 0.15.2 -n metallb-system --create-namespace
kubectl wait --timeout=5m -n metallb-system deployment/metallb-controller --for=condition=Available
kubectl apply -f "$REPO_DIR/k8s/metallb/"

# Install and Configure Knative Gateway
kubectl apply -f "$REPO_DIR/k8s/gateway/"
kubectl wait --timeout=5m -n gateway gateway/knative-gateway --for=condition=Programmed

# Install and Configure Knative Operator
helm repo add knative-operator https://knative.github.io/operator
helm upgrade --install knative-operator knative-operator/knative-operator --version 1.18.0 -n knative-operator --create-namespace
kubectl wait --timeout=5m -n knative-operator deployment/knative-operator --for=condition=Available

kubectl apply -f "$REPO_DIR/k8s/knative/01-ns.yaml"
kubectl apply -f "$REPO_DIR/k8s/knative/02-knative-serving.yaml"
kubectl wait --timeout=5m -n knative-serving deployment/controller --for=condition=Available

kubectl apply -f "$REPO_DIR/k8s/knative/03-net-gateway-api.yaml"
kubectl wait --timeout=5m -n knative-serving deployment/net-gateway-api-controller --for=condition=Available

# Install our Workload
kubectl apply -f "$REPO_DIR/k8s/workload/"


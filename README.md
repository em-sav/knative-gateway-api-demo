# CNCF Presentation - Knative Serving with Gateway API plugin

[![gateway api](https://img.shields.io/badge/gateway_api-326ce5)](https://gateway-api.sigs.k8s.io/)
[![istio](https://img.shields.io/badge/istio-516ba9)](https://istio.io/latest/)
[![metallb](https://img.shields.io/badge/metallb-1d90f3)](https://metallb.io/)
[![kind](https://img.shields.io/badge/KinD-02b7a5)](https://kind.sigs.k8s.io/)
[![knative](https://img.shields.io/badge/knative-0366ad)](https://kind.sigs.k8s.io/)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/knative/serving)


A complete [Knative Serving](https://knative.dev/docs/serving/) demo with [Gateway API](https://gateway-api.sigs.k8s.io/) plugin. This README guides you through setting up a local Kubernetes cluster using [KinD](https://kind.sigs.k8s.io/) and deploying Knative Serving with Gateway API.

[![Knative Presentation (French)](https://i.ytimg.com/vi/81Mpg6mI__0/hqdefault.jpg)](https://www.youtube.com/watch?v=81Mpg6mI__0&t=4897s)

## 📝 Prerequisites

Required CLIs:
- `docker`
- `kind`
- `kubectl`
- `helm`
- `go`

### Update `/etc/hosts`

Add the following entry to your `hosts` file:

```
# /etc/hosts
...
172.18.254.254  knative-serving.cncf.demo
```

If you can't edit the `/etc/hosts` file, add `--resolve knative-serving.cncf.demo:80:172.18.254.254` to your **curl** commands.

### MacOS users

This enables direct connectivity from your laptop to the KinD cluster node subnet (required for MetalLB IP allocation).

```bash
brew install chipmk/tap/docker-mac-net-connect

sudo docker-mac-net-connect
```

## Demo

### ☸️ Bootstrap KinD cluster

```bash
./cluster/bootstrap.sh
```

Takes around 1 minute to fully configure the cluster.

### 💻 Scenarios

#### 1. Simple Knative Serving deployment

```bash
kubectl apply -f ./workload/01-cncf-server.yaml

curl -v http://knative-serving.cncf.demo/status/200
```

#### 2. Traffic splitting between two revisions (Canary deployment)

- 90% of traffic to revision v1
- 10% of traffic to revision v2

```bash
kubectl apply -f ./workload/02-cncf-server.yaml

go run load-test/main.go -url http://knative-serving.cncf.demo/status/201 -n 5000 -c 200
```

#### 3. Full rollout of new revision

- 100% of traffic to revision v2

```bash
kubectl apply -f ./workload/03-cncf-server.yaml

go run load-test/main.go -url http://knative-serving.cncf.demo/status/201 -n 5000 -c 200
```

### 🧹 Cleanup

```bash
./cluster/cleanup.sh
```

![CNCF](https://www.cncf.io/wp-content/uploads/2022/07/cncf-color-bg.svg)

#!/bin/bash

# VPS Setup Script untuk k3s + MetalLB + Traefik
# Jalankan di VPS 72.61.69.116

set -e

echo "🚀 Setting up k3s cluster on VPS..."

# Install k3s dengan Traefik disabled (kita akan configure manual)
curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--disable traefik" sh -

# Wait for k3s to be ready
echo "⏳ Waiting for k3s to be ready..."
sleep 30

# Install MetalLB
echo "📦 Installing MetalLB..."
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.12/config/manifests/metallb-native.yaml

# Wait for MetalLB to be ready
echo "⏳ Waiting for MetalLB to be ready..."
kubectl wait --namespace metallb-system \
                --for=condition=ready pod \
                --selector=app=metallb \
                --timeout=90s

# Apply MetalLB config
echo "⚙️  Configuring MetalLB..."
kubectl apply -f - <<EOF
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: default-pool
  namespace: metallb-system
spec:
  addresses:
  - 72.61.69.116/32
---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: default
  namespace: metallb-system
spec:
  ipAddressPools:
  - default-pool
EOF

# Install Traefik with proper config
echo "🌐 Installing Traefik..."
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: traefik-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: traefik
  namespace: traefik-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: traefik
rules:
- apiGroups: [""]
  resources: ["services","endpoints","secrets"]
  verbs: ["get","list","watch"]
- apiGroups: ["extensions","networking.k8s.io"]
  resources: ["ingresses","ingressclasses"]
  verbs: ["get","list","watch"]
- apiGroups: ["extensions","networking.k8s.io"]
  resources: ["ingresses/status"]
  verbs: ["update"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: traefik
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: traefik
subjects:
- kind: ServiceAccount
  name: traefik
  namespace: traefik-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: traefik
  namespace: traefik-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: traefik
  template:
    metadata:
      labels:
        app: traefik
    spec:
      serviceAccountName: traefik
      containers:
      - name: traefik
        image: traefik:v3.0
        args:
        - --entrypoints.web.address=:80
        - --entrypoints.websecure.address=:443
        - --entrypoints.web.http.redirections.entrypoint.to=websecure
        - --entrypoints.web.http.redirections.entrypoint.scheme=https
        - --providers.kubernetesingress=true
        - --certificatesresolvers.letsencrypt.acme.httpchallenge=true
        - --certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web
        - --certificatesresolvers.letsencrypt.acme.email=admin@amconsole.store
        - --certificatesresolvers.letsencrypt.acme.storage=/data/acme.json
        - --log.level=INFO
        ports:
        - containerPort: 80
        - containerPort: 443
        volumeMounts:
        - name: acme
          mountPath: /data
      volumes:
      - name: acme
        hostPath:
          path: /var/lib/traefik
          type: DirectoryOrCreate
---
apiVersion: v1
kind: Service
metadata:
  name: traefik
  namespace: traefik-system
spec:
  type: LoadBalancer
  loadBalancerIP: 72.61.69.116
  selector:
    app: traefik
  ports:
  - name: web
    port: 80
    targetPort: 80
  - name: websecure
    port: 443
    targetPort: 443
---
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: traefik
spec:
  controller: traefik.io/ingress-controller
EOF

echo "✅ k3s cluster setup completed!"
echo "📋 Next steps:"
echo "1. Copy kubeconfig: sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config"
echo "2. Fix kubeconfig server IP: sed -i 's/127.0.0.1/72.61.69.116/g' ~/.kube/config"
echo "3. Test: kubectl get nodes"
echo "4. Deploy your app: kubectl apply -f k8s/"
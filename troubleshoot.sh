#!/bin/bash

# Troubleshooting script untuk k3s cluster

echo "🔍 Checking k3s cluster status..."

echo "📊 Cluster Info:"
kubectl cluster-info

echo ""
echo "🖥️  Nodes:"
kubectl get nodes -o wide

echo ""
echo "🏠 Namespaces:"
kubectl get namespaces

echo ""
echo "🌐 Traefik Status:"
kubectl get pods,svc -n traefik-system

echo ""
echo "⚖️  MetalLB Status:"
kubectl get pods,svc -n metallb-system

echo ""
echo "📱 Application Status:"
kubectl get all -n amconsole

echo ""
echo "🔗 Ingress:"
kubectl get ingress -n amconsole

echo ""
echo "📜 Recent Events:"
kubectl get events --sort-by=.metadata.creationTimestamp -n amconsole | tail -10

echo ""
echo "🔍 Traefik Logs (last 20 lines):"
kubectl logs -n traefik-system deployment/traefik --tail=20

echo ""
echo "🌍 External IP Status:"
kubectl get svc -n traefik-system traefik